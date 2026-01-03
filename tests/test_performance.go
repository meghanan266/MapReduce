package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "time"
)

type SplitResponse struct {
    ChunkURLs []string `json:"chunk_urls"`
}

type MapResponse struct {
    ResultURL string `json:"result_url"`
}

type ReduceResponse struct {
    FinalResultURL string `json:"final_result_url"`
}

func main() {
    // Your service IPs
    splitterIP := "54.158.42.116"
    mapperIPs := []string{"54.147.211.79", "54.163.78.164", "35.175.212.209"}
    reducerIP := "100.25.136.37"
    
    bucket := "mapreduce-meghana-20250928"
    file := "input.txt"
    
    fmt.Println("Starting MapReduce Performance Test")
    fmt.Println("=====================================")
    
    totalStart := time.Now()
    
    // Step 1: Split
    splitStart := time.Now()
    splitURL := fmt.Sprintf("http://%s:8080/split?bucket=%s&file=%s", splitterIP, bucket, file)
    resp, err := http.Get(splitURL)
    if err != nil {
        fmt.Printf("Split failed: %v\n", err)
        return
    }
    
    body, _ := ioutil.ReadAll(resp.Body)
    resp.Body.Close()
    
    var splitResp SplitResponse
    json.Unmarshal(body, &splitResp)
    
    splitTime := time.Since(splitStart)
    fmt.Printf("Split Time: %v\n", splitTime)
    fmt.Printf("Chunks created: %d\n", len(splitResp.ChunkURLs))
    
    // Step 2: Map
    mapStart := time.Now()
    var mapResults []string
    
    for i, chunkURL := range splitResp.ChunkURLs {
        mapperIP := mapperIPs[i%len(mapperIPs)]
        mapURL := fmt.Sprintf("http://%s:8080/map?chunk_url=%s", mapperIP, chunkURL)
        
        mapResp, err := http.Get(mapURL)
        if err != nil {
            fmt.Printf("Map %d failed: %v\n", i, err)
            continue
        }
        
        mapBody, _ := ioutil.ReadAll(mapResp.Body)
        mapResp.Body.Close()
        
        var mapResponse MapResponse
        json.Unmarshal(mapBody, &mapResponse)
        mapResults = append(mapResults, mapResponse.ResultURL)
        
        fmt.Printf("Mapper %d completed\n", i+1)
    }
    
    mapTime := time.Since(mapStart)
    fmt.Printf("Total Map Time: %v\n", mapTime)
    
    // Step 3: Reduce
    reduceStart := time.Now()
    reduceURL := fmt.Sprintf("http://%s:8080/reduce", reducerIP)
    
    reduceResp, err := http.Get(reduceURL)
    if err != nil {
        fmt.Printf("Reduce failed: %v\n", err)
        return
    }
    
    reduceBody, _ := ioutil.ReadAll(reduceResp.Body)
    reduceResp.Body.Close()
    
    var reduceResponse ReduceResponse
    json.Unmarshal(reduceBody, &reduceResponse)
    
    reduceTime := time.Since(reduceStart)
    fmt.Printf("Reduce Time: %v\n", reduceTime)
    
    totalTime := time.Since(totalStart)
    
    fmt.Println("\n=== Performance Summary ===")
    fmt.Printf("Split Time:  %v\n", splitTime)
    fmt.Printf("Map Time:    %v\n", mapTime)
    fmt.Printf("Reduce Time: %v\n", reduceTime)
    fmt.Printf("Total Time:  %v\n", totalTime)
    fmt.Printf("Final Result: %s\n", reduceResponse.FinalResultURL)
}