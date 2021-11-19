package main

import (
	"fmt"
	"os"
	"time"

	"github.com/apoorvam/goterminal"
	"github.com/cavaliercoder/grab"
)

func (gnvm *GNVM) DownloadBinaryFile(version string) {
	gnvm.CurrentState <- STARTED_DOWNLOADING

	client := grab.NewClient()
	req, _ := grab.NewRequest(".", fmt.Sprintf("%s/%s", NODE_VERSIONS_URL, version))

	// start download
	// fmt.Printf("Downloading %v...\n", req.URL())
	resp := client.Do(req)
	// fmt.Printf("  %v\n", resp.HTTPResponse.Status)

	// start UI loop
	t := time.NewTicker(200 * time.Millisecond)
	defer t.Stop()

	writer := goterminal.New(os.Stdout)
Loop:
	for {
		select {
		case <-t.C:
			writer.Clear()

			fmt.Fprintf(
				writer, "Downloading (%d/%d) bytes [%.2f%%]\n",
				resp.BytesComplete(),
				resp.Size,
				100*resp.Progress(),
			)

			writer.Print()
		case <-resp.Done:
			break Loop
		}
	}

	writer.Reset()

	if err := resp.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
		os.Exit(1)
	}

	// get the file work on it then delete it on finishing

	gnvm.binaryDownloadName = resp.Filename
	gnvm.CurrentState <- FINISHED_DOWNLOADING
}
