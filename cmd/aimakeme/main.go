package main

import (
	"aimakeme"
	"flag"
	"fmt"
	"os"
	"sync"
)

func main() {
	opts := aimakeme.Options{}
	opts.Folder = flag.String("f", "default", "Folder, a way to bundle multiple runs in the same folder")
	opts.N = flag.Int("n", 1, "Number, how many to generate at once")
	opts.Prompt = flag.String("p", "An image that makes you happy", "Prompt, what to draw")
	opts.Style = flag.String("s", "vivid", "Style, either vivid or natural")
	opts.APIKey = os.Getenv("OPENAI_APIKEY")

	flag.Parse()
	opts.Wait = &sync.WaitGroup{}
	opts.Wait.Add(*opts.N)

	if err := aimakeme.Run(opts); err != nil {
		fmt.Println(err)
	}
	opts.Wait.Wait()
}
