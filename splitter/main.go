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

type SplitResponse struct {
    ChunkURLs []string `json:"chunk_urls"`
}

func main() {
    http.HandleFunc("/", handler)
    http.HandleFunc("/split", splitHandler)
    
    port := ":8080"
    fmt.Printf("Splitter service starting on port %s\n", port)
    log.Fatal(http.ListenAndServe(port, nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Splitter service is running!")
}

func splitHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Only GET method allowed", http.StatusMethodNotAllowed)
        return
    }
    
    // Get query parameters
    bucketName := r.URL.Query().Get("bucket")
    fileName := r.URL.Query().Get("file")
    
    if bucketName == "" || fileName == "" {
        bucketName = "mapreduce-meghana-20250928"
        fileName = "input.txt"
    }
    
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
    
    // Download file from S3
    result, err := svc.GetObject(&s3.GetObjectInput{
        Bucket: aws.String(bucketName),
        Key:    aws.String(fileName),
    })
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to get object: %v", err), http.StatusInternalServerError)
        return
    }
    defer result.Body.Close()
    
    // Read the file content
    content, err := ioutil.ReadAll(result.Body)
    if err != nil {
        http.Error(w, "Failed to read file content", http.StatusInternalServerError)
        return
    }
    
    // Convert to string and split into 3 chunks
    text := string(content)
    chunkSize := len(text) / 3
    
    chunks := []string{
        text[0:chunkSize],
        text[chunkSize:chunkSize*2],
        text[chunkSize*2:],
    }
    
    // Upload chunks back to S3
    var chunkURLs []string
    for i, chunk := range chunks {
        chunkKey := fmt.Sprintf("chunks/%s_chunk_%d.txt", strings.TrimSuffix(fileName, ".txt"), i)
        
        _, err := svc.PutObject(&s3.PutObjectInput{
            Bucket: aws.String(bucketName),
            Key:    aws.String(chunkKey),
            Body:   bytes.NewReader([]byte(chunk)),
        })
        if err != nil {
            http.Error(w, fmt.Sprintf("Failed to upload chunk: %v", err), http.StatusInternalServerError)
            return
        }
        
        chunkURL := fmt.Sprintf("s3://%s/%s", bucketName, chunkKey)
        chunkURLs = append(chunkURLs, chunkURL)
    }
    
    // Return the chunk URLs
    response := SplitResponse{
        ChunkURLs: chunkURLs,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}