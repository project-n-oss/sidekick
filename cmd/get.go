package cmd

import (
	"net/http"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use: "get",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := http.Get("http://sdk-test-rvh.localhost:8000/data.csv?x-id=GetObject")
		logger.Fatal(err.Error())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
