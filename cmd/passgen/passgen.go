package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var chars = []rune(
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")

type userInfo struct {
	Password string `json:"password"`
	Role     string `json:"role"`
}

func main() {
	usersList := flag.String("i", "users.txt", "list of usernames")
	credsList := flag.String("o", ".creds.json", "list of users login credentials")
	flag.Parse()

	users := make(map[string]*userInfo)
	users["admin"] = &userInfo{
		Role:     "admin",
		Password: "VguRobocon@2019",
	}

	input, err := os.Open(*usersList)
	if err != nil {
		log.Fatalf("could no open input file: %v", err)
	}

	output, err := os.Create(*credsList)
	if err != nil {
		log.Fatalf("could no open output file: %v", err)
	}

	scanner := bufio.NewScanner(input)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		username := string(scanner.Bytes())
		if _, ok := users[username]; ok {
			log.Errorf("user exists! Skipping")
			continue
		}

		info := &userInfo{
			Role:     "contestant",
			Password: genPass(8),
		}
		users[username] = info
	}

	b, err := json.MarshalIndent(users, "", "\t")
	if err != nil {
		log.Fatalf("could no unmarshal json: %v", err)
	}

	w := bufio.NewWriter(output)
	n, err := w.Write(b)
	if err != nil {
		log.Fatalf("could not write to files: %v", err)
	}
	w.Flush()

	fmt.Printf("wrote %d bytes\n", n)
}

func genPass(n int) string {
	rand.Seed(time.Now().UnixNano())
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	str := b.String()

	return str
}
