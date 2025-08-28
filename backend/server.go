// package main

// import "github.com/gin-gonic/gin"

// func main() {
// 	server := gin.Default()

// 	server.GET("/test", func(ctx *gin.Context) {
// 		ctx.JSON(200, gin.H{
// 			"message": "cool",
// 		})
// 	})

// 	server.Run("127.0.0.1:8080")
// }

package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Response struct {
	Message string `json:"message"`
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp := Response{
		Message: "cool",
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func main() {
	http.HandleFunc("/test", testHandler)
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", nil))
}
