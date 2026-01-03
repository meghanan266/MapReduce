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

type ReduceRequest struct {
    MapperURLs []string `json:"mapper_urls"`
}

type ReduceResponse struct {
    FinalResultURL string `json:"final_result_url"`
}

func main() {
    http.HandleFunc("/", handler)
    http.HandleFunc("/reduce", reduceHandler)
    
    port := ":8080"
    fmt.Printf("Reducer service starting on port %s\n", port)
    log.Fatal(http.ListenAndServe(port, nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Reducer service is running!")
}

func reduceHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Only GET method allowed", http.StatusMethodNotAllowed)
        return
    }
    
    // Get mapper URLs from query parameters
    mapperURLs := r.URL.Query()["mapper_url"]
    if len(mapperURLs) == 0 {
        // Default to the 3 mapped files if no URLs provided
        mapperURLs = []string{
            "s3://mapreduce-meghana-20250928/mapped/input_chunk_0_mapped.json",
            "s3://mapreduce-meghana-20250928/mapped/input_chunk_1_mapped.json",
            "s3://mapreduce-meghana-20250928/mapped/input_chunk_2_mapped.json",
        }
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
    
    // Aggregate all word counts
    finalWordCount := make(map[string]int)
    
    for _, mapperURL := range mapperURLs {
        // Parse S3 URL
        parts := strings.TrimPrefix(mapperURL, "s3://")
        splitParts := strings.SplitN(parts, "/", 2)
        if len(splitParts) != 2 {
            continue
        }
        bucketName := splitParts[0]
        key := splitParts[1]
        
        // Download mapper result from S3
        result, err := svc.GetObject(&s3.GetObjectInput{
            Bucket: aws.String(bucketName),
            Key:    aws.String(key),
        })
        if err != nil {
            http.Error(w, fmt.Sprintf("Failed to get mapper result: %v", err), http.StatusInternalServerError)
            return
        }
        
        // Read content
        content, err := ioutil.ReadAll(result.Body)
        result.Body.Close()
        if err != nil {
            http.Error(w, "Failed to read mapper result", http.StatusInternalServerError)
            return
        }
        
        // Parse JSON word count
        var wordCount map[string]int
        if err := json.Unmarshal(content, &wordCount); err != nil {
            http.Error(w, "Failed to parse mapper result", http.StatusInternalServerError)
            return
        }
        
        // Merge counts
        for word, count := range wordCount {
            finalWordCount[word] += count
        }
    }
    
    // Convert to JSON
    jsonData, err := json.Marshal(finalWordCount)
    if err != nil {
        http.Error(w, "Failed to marshal final result", http.StatusInternalServerError)
        return
    }
    
    // Upload final result to S3
    resultKey := "final/word_count_final.json"
    bucketName := "mapreduce-meghana-20250928"
    
    _, err = svc.PutObject(&s3.PutObjectInput{
        Bucket: aws.String(bucketName),
        Key:    aws.String(resultKey),
        Body:   bytes.NewReader(jsonData),
    })
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to upload final result: %v", err), http.StatusInternalServerError)
        return
    }
    
    // Return final result URL
    response := ReduceResponse{
        FinalResultURL: fmt.Sprintf("s3://%s/%s", bucketName, resultKey),
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}