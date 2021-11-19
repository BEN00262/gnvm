package main

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"go.etcd.io/bbolt"
)

func main() {
	db, err := bbolt.Open("gnvm", 0600, nil)

	if err != nil {
		log.Fatal(err.Error())
	}

	defer db.Close()

	gnvm := NewGNVM(db)
	defer gnvm.wg.Wait()

	// start listening for the commands

	rootCmd := &cobra.Command{
		Use:   "gnvm",
		Short: "Install nodejs versions",
	}

	// cobra configuration
	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install nodejs versions",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(args)
			if err := gnvm.InstallNodeJSVersion(args[0]); err != nil {
				log.Fatal(err.Error())
			}
		},
	}

	listLocal := &cobra.Command{
		Use:   "ls-local",
		Short: "ls nodejs versions",
		Run: func(cmd *cobra.Command, args []string) {
			if err := gnvm.ListLocalNodeVersions(); err != nil {
				log.Fatal(err.Error())

				db.Close()
			}
		},
	}

	listOnTheInternet := &cobra.Command{
		Use:   "ls",
		Short: "ls nodejs versions",
		Run: func(cmd *cobra.Command, args []string) {
			if err := gnvm.GetAllNodeVersions(func(nj []NodeJS) {
				for _, node := range nj {
					println(node.Version)
				}
			}); err != nil {
				log.Fatal(err.Error())
			}
		},
	}

	listOnTheInternet.AddCommand(
		&cobra.Command{
			Use:   "refresh",
			Short: "ls nodejs versions",
			Run: func(cmd *cobra.Command, args []string) {
				if err := gnvm.GetAllNodeVersions(func(nj []NodeJS) {
					for _, node := range nj {
						println(node.Version)
					}
				}); err != nil {
					log.Fatal(err.Error())
				}
			},
		},
	)

	rootCmd.AddCommand(
		installCmd,
		listLocal,
		listOnTheInternet,
	)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}
