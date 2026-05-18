package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/Future-Game-Laboratory/Steins-Gate/dbschema"
	"github.com/Future-Game-Laboratory/Steins-Gate/mysql"
)

func main() {
	reset := flag.Bool("reset", true, "drop existing tables before creating them")
	timeout := flag.Duration("timeout", 15*time.Second, "database init timeout")
	flag.Parse()

	if err := mysql.Init(); err != nil {
		log.Fatal(err)
	}
	defer mysql.Close()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	if *reset {
		if err := dbschema.Reset(ctx, mysql.GetDB()); err != nil {
			log.Fatal(err)
		}
		log.Println("database tables dropped and recreated")
		return
	}

	if err := dbschema.Ensure(ctx, mysql.GetDB()); err != nil {
		log.Fatal(err)
	}
	log.Println("database tables ensured")
}
