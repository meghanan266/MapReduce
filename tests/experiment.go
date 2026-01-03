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

func runExperiment(splitterIP string, mapperIPs []string, reducerIP string, bucket string, file string) (float64, float64, float64, float64) {
    totalStart := time.Now()
    
    // Split
    splitStart := time.Now()
    splitURL := fmt.Sprintf("http://%s:8080/split?bucket=%s&file=%s", splitterIP, bucket, file)
    resp, _ := http.Get(splitURL)
    body, _ := ioutil.ReadAll(resp.Body)
    resp.Body.Close()
    var splitResp SplitResponse
    json.Unmarshal(body, &splitResp)
    splitTime := time.Since(splitStart).Seconds()
    
    // Map - Parallel
    mapStart := time.Now()
    for i, chunkURL := range splitResp.ChunkURLs {
        mapperIP := mapperIPs[i%len(mapperIPs)]
        mapURL := fmt.Sprintf("http://%s:8080/map?chunk_url=%s", mapperIP, chunkURL)
        mapResp, _ := http.Get(mapURL)
        mapResp.Body.Close()
    }
    mapTime := time.Since(mapStart).Seconds()
    
    // Reduce
    reduceStart := time.Now()
    reduceURL := fmt.Sprintf("http://%s:8080/reduce", reducerIP)
    reduceResp, _ := http.Get(reduceURL)
    reduceResp.Body.Close()
    reduceTime := time.Since(reduceStart).Seconds()
    
    totalTime := time.Since(totalStart).Seconds()
    
    return splitTime, mapTime, reduceTime, totalTime
}

func main() {
    splitterIP := "54.158.42.116"
    mapperIPs := []string{"54.147.211.79", "54.163.78.164", "35.175.212.209"}
    reducerIP := "100.25.136.37"
    bucket := "mapreduce-meghana-20250928"
    file := "input.txt"
    
    fmt.Println("Running Performance Experiments")
    fmt.Println("================================")
    
    // Experiment 1: Run 5 times to show consistency
    fmt.Println("\nExperiment 1: Consistency Test (5 runs)")
    fmt.Println("Run\tSplit\tMap\tReduce\tTotal")
    
    var avgSplit, avgMap, avgReduce, avgTotal float64
    
    for i := 1; i <= 5; i++ {
        split, mapTime, reduce, total := runExperiment(splitterIP, mapperIPs, reducerIP, bucket, file)
        fmt.Printf("%d\t%.3f\t%.3f\t%.3f\t%.3f\n", i, split, mapTime, reduce, total)
        
        avgSplit += split
        avgMap += mapTime
        avgReduce += reduce
        avgTotal += total
        
        time.Sleep(2 * time.Second) // Wait between runs
    }
    
    avgSplit /= 5
    avgMap /= 5
    avgReduce /= 5
    avgTotal /= 5
    
    fmt.Printf("\nAverages:\n")
    fmt.Printf("Split: %.3f sec\n", avgSplit)
    fmt.Printf("Map: %.3f sec\n", avgMap)
    fmt.Printf("Reduce: %.3f sec\n", avgReduce)
    fmt.Printf("Total: %.3f sec\n", avgTotal)
    
    // Calculate speedup (assuming sequential would take 3x map time)
    sequentialEstimate := avgSplit + (avgMap * 3) + avgReduce
    speedup := sequentialEstimate / avgTotal
    
    fmt.Printf("\nEstimated Sequential Time: %.3f sec\n", sequentialEstimate)
    fmt.Printf("Parallel Time: %.3f sec\n", avgTotal)
    fmt.Printf("Speedup: %.2fx\n", speedup)
}