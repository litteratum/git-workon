package main

import (
	"encoding/json"
	"log"
	"os"
	"path"
	"path/filepath"
	"syscall"
)

var cachePath = path.Join(getCacheDir(), "projects.json")

type ICache interface {
	Get(project string) ProjectInfo
	Set(project string, info ProjectInfo)
	Write()
}

type ProjectInfo struct {
	Source string `json:"source"`
}

type Cache struct {
	Data map[string]ProjectInfo
}

func (c Cache) Get(project string) ProjectInfo {
	return c.Data[project]
}

func (c Cache) Set(project string, info ProjectInfo) {
	c.Data[project] = info
}

func (c Cache) Write() {
	data, err := json.MarshalIndent(c.Data, "", "  ")
	if err != nil {
		log.Printf("failed to marshal the cache: %s", err)
	}

	err = os.WriteFile(cachePath, data, 0644)
	if err != nil {
		log.Printf("failed to write the cache file at %s: %s", cachePath, err)
	}
}

func NewCache(data map[string]ProjectInfo) *Cache {
	return &Cache{
		Data: data,
	}
}

func NewCacheFromFile() *Cache {
	data, err := os.ReadFile(cachePath)
	if os.IsNotExist(err) {
		return createCacheFile()
	}

	if err != nil {
		log.Fatalf("failed to read the cache file at %s: %s", cachePath, err)
	}

	var data_ map[string]ProjectInfo
	err = json.Unmarshal(data, &data_)
	if err != nil {
		log.Fatalf("failed to unmarshal the cache file at %s: %s", cachePath, err)
	}

	return NewCache(data_)
}

func getCacheDir() string {
	systemDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatalf("failed to get the user cache directory: %s", err)
	}

	return path.Join(systemDir, "git_workon")
}

func createCacheFile() *Cache {
	data := map[string]ProjectInfo{}

	dir := filepath.Dir(cachePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Printf("failed to create cache dirs %s: %s", dir, err)
	}

	f, err := os.OpenFile(cachePath, os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if pe, ok := err.(*os.PathError); ok && pe.Err == syscall.EEXIST {
			return nil
		}
		log.Printf("failed to create cache file %s: %s", cachePath, err)
	}
	defer f.Close()

	return NewCache(data)
}
