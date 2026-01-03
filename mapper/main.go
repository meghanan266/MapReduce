package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "strings"
    
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
)

type MapResponse struct {
    ResultURL string `json:"result_url"`
}

func main() {
    http.HandleFunc("/", handler)
    http.HandleFunc("/map", mapHandler)
    
    port := ":8080"
    fmt.Printf("Mapper service starting on port %s\n", port)
    log.Fatal(http.ListenAndServe(port, nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Mapper service is running!")
}

func mapHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Only GET method allowed", http.StatusMethodNotAllowed)
        return
    }
    
    // Get S3 URL from query parameter
    s3URL := r.URL.Query().Get("chunk_url")
    if s3URL == "" {
        http.Error(w, "chunk_url parameter required", http.StatusBadRequest)
        return
    }
    
    // Parse S3 URL (format: s3://bucket/key)
    parts := strings.TrimPrefix(s3URL, "s3://")
    splitParts := strings.SplitN(parts, "/", 2)
    if len(splitParts) != 2 {
        http.Error(w, "Invalid S3 URL format", http.StatusBadRequest)
        return
    }
    bucketName := splitParts[0]
    key := splitParts[1]
    
    // Create AWS session
    sess, err := session.NewSession(&aws.Config{
        Region: aws.String("us-east-1"),
    })
    if err != nil {
        http.Error(w, "Failed to create AWS session", http.StatusInternalServerError)
        return
    }
    
    // Create S3 client
    svc := s3.New(sess)
    
    // Download chunk from S3
    result, err := svc.GetObject(&s3.GetObjectInput{
        Bucket: aws.String(bucketName),
        Key:    aws.String(key),
    })
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to get object: %v", err), http.StatusInternalServerError)
        return
    }
    defer result.Body.Close()
    
    // Read content
    content, err := ioutil.ReadAll(result.Body)
    if err != nil {
        http.Error(w, "Failed to read content", http.StatusInternalServerError)
        return
    }
    
    // Count words
    text := string(content)
    words := strings.Fields(strings.ToLower(text))
    wordCount := make(map[string]int)
    
    for _, word := range words {
        // Remove punctuation
        word = strings.Trim(word, ".,!?;:\"'")
        if word != "" {
            wordCount[word]++
        }
    }
    
    // Convert to JSON
    jsonData, err := json.Marshal(wordCount)
    if err != nil {
        http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
        return
    }
    
    // Upload result to S3
    resultKey := strings.Replace(key, "chunks/", "mapped/", 1)
    resultKey = strings.Replace(resultKey, ".txt", "_mapped.json", 1)
    
    _, err = svc.PutObject(&s3.PutObjectInput{
        Bucket: aws.String(bucketName),
        Key:    aws.String(resultKey),
        Body:   bytes.NewReader(jsonData),
    })
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to upload result: %v", err), http.StatusInternalServerError)
        return
    }
    
    // Return result URL
    response := MapResponse{
        ResultURL: fmt.Sprintf("s3://%s/%s", bucketName, resultKey),
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}