package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func p2p_func(c *gin.Context) {

	if start_p2p_flag == 1 {
		return
	} else {
		go start_p2p()
		start_p2p_flag = 1
	}
	c.Status(http.StatusOK)
}
func test_connection(c *gin.Context) {
	var TEST test_struct
	c.BindJSON(&TEST)
	peer_details_list = strings.Split(TEST.Peer_list, " ")

	test_conn() //Adds host ip and connects to all
	time.Sleep(time.Second * 3)

	c.JSON(http.StatusOK, gin.H{
		"All ok": all_ok,
	})
}

func id_func(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"Host IP": p2p.Host_ip, //.Host.ID()),
	})
	// return
}
func defaults() {
	status_struct.Chan = "81247"
	start_p2p_flag = 0

}
func gen_keyshares(c *gin.Context) {
	test_conn() //Adds host ip and connects to all
	time.Sleep(time.Second * 3)
	//Send back data to send to cloud

	c.JSON(http.StatusOK, gin.H{
		"All ok": all_ok,
	})
}

func msg_sign(c *gin.Context) {
	test_conn() //Adds host ip and connects to all
	time.Sleep(time.Second * 3)
	//Sign part signatre

	c.JSON(http.StatusOK, gin.H{
		"All ok": all_ok,
	})
}

func main() {

	defaults()
	p2p = *start_p2p()
	if debug == true {
		local()

	} else {

		router := gin.Default()
		router.Use(cors.New(cors.Config{
			// AllowOrigins:    []string{"http://localhost:8080", "http://127.0.0.1:3000"},
			AllowMethods:     []string{"POST", "GET"},
			AllowHeaders:     []string{"Origin"},
			AllowAllOrigins:  true,
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}))
		v1 := router.Group("/api")
		{
			v1.GET("/start_p2p", p2p_func)
			v1.GET("/get_ID", id_func)
			v1.POST("/test_connection", test_connection)
			v1.POST("/gen_keyshares", gen_keyshares)
			v1.POST("/sign_message", msg_sign)

		}
		router.Run(":8070")
	}

}
