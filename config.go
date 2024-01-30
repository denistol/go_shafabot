package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"time"
)

type ProductData struct {
	Timestamp int64  `json:"timestamp"`
	Url       string `json:"url"`
}

type ProductsMap map[int]ProductData

const CONFIG_PATH string = "./config.json"
const LIFETIME int64 = 12 * 60 * 60 // seconds

type Config struct {
	Timeout int64       `json:"timeout"`
	Token   string      `json:"token"`
	ChatIds []int64     `json:"chat_ids"`
	Urls    []string    `json:"urls"`
	Log     ProductsMap `json:"log"`
}

func (c *Config) removeOldProducts() {
	var now = time.Now().Unix()

	for k, v := range c.Log {
		var diff = now - v.Timestamp
		if diff >= LIFETIME {
			delete(c.Log, k)
		}
	}
	c.save()
}

func (c *Config) handleProducts(incoming ProductsMap) ProductsMap {
	var hasKeys = len(c.Log) != 0
	var uniq = make(ProductsMap)

	for k, v := range incoming {
		_, exists := c.Log[k]
		if !exists && hasKeys {
			uniq[k] = v
		}
		c.Log[k] = v
	}
	c.save()
	return uniq
}

func (c *Config) save() {
	file, err := json.MarshalIndent(c, "", " ")
	if err == nil {
		os.WriteFile(CONFIG_PATH, file, fs.ModeCharDevice)
	}
}

func ConfigConstructor() Config {
	var payload Config
	content, err := os.ReadFile(CONFIG_PATH)
	if err != nil {
		fmt.Print(err)
	}

	err = json.Unmarshal(content, &payload)
	if err != nil {
		log.Fatal("Error getConfig: ", err)
	}
	return payload
}
