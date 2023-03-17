package cmd

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use: "get",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := http.Get("http://sdk-test-rvh.localhost:8000/data.csv?x-id=GetObject")
		if err != nil {
			logger.Fatal(err.Error())
		}
		fmt.Printf("GOT: %v\n", resp.StatusCode)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
