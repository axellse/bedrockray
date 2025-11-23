package main

import (
	"fmt"
	"net/http"
)

func StartDashboard(port string, sc ServerController) {
	fmt.Println("now starting dashboard")
	http.ListenAndServe(":" + port, nil)
}