package utils

import (
	"github.com/bwmarrin/snowflake"
	"math/rand"
	"time"
)

//***************************************************
//author tim.huang
//2022/12/27
//
//
//***************************************************

//***********************    type    ****************************

//***********************    type_end    ****************************

//***********************    var    ****************************

//***********************    var_end    ****************************

//***********************    interface    ****************************

//***********************    interface_end    ****************************

//***********************    struct    ****************************

//***********************    struct_end    ****************************

func init() {
	//初始化雪花算法Id
	source := rand.NewSource(time.Now().UnixNano())
	ran := rand.New(source)
	Snowflake, _ = snowflake.NewNode(ran.Int63n(1024))
}
