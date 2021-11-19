package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	bolt "go.etcd.io/bbolt"
)

const (
	INSTALLED_BUCKET    = "installed_node_versions"
	ALL_VERSIONS_BUCKET = "all_node_versions"
	NODE_VERSIONS_URL   = "https://nodejs.org/en/download/releases/"
)

type NodeJS struct {
	Version string
	Link    string
}

type STATE int

const (
	FINISHED_DOWNLOADING STATE = iota + 1
	STARTED_DOWNLOADING
)

type GNVM struct {
	CurrentState       chan STATE
	Versions           []NodeJS
	Db                 *bolt.DB
	wg                 *sync.WaitGroup
	binaryDownloadName string
}

func NewGNVM(db *bolt.DB) *GNVM {
	return &GNVM{
		CurrentState: make(chan STATE, 1),
		Db:           db,
		wg:           &sync.WaitGroup{},
	}
}

// store this data in a db then use it later
// pass the values to a channel to be consumed later
func (gnvm *GNVM) ListLocalNodeVersions() error {
	return gnvm.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(INSTALLED_BUCKET))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var nodejs_version NodeJS

			if err := json.Unmarshal(v, &nodejs_version); err != nil {
				log.Fatal(err.Error())
			}

			fmt.Println(nodejs_version.Version)
		}

		return nil
	})
}

// get a version string ( check it against the online list and the offline list )
func (gnvm *GNVM) InstallNodeJSVersion(version string) error {
	// get the version specified then install it
	return gnvm.Db.View(func(tx *bolt.Tx) error {
		value := tx.Bucket([]byte(ALL_VERSIONS_BUCKET)).Get([]byte(version))

		if value != nil {
			var nodejs_version NodeJS

			if err := json.Unmarshal(value, &nodejs_version); err != nil {
				return err
			}

			// we need to download the binary first before proceeding
			gnvm.wg.Add(1)
			go gnvm.DownloadBinaryFile(GetNodeFileVersion(nodejs_version.Version))

			// we should start monitoring the states until state changes to either finshed or not
			for state := range gnvm.CurrentState {
				switch state {
				case STARTED_DOWNLOADING:
					{
						// we dont care about this
					}
				case FINISHED_DOWNLOADING:
					{
						// do the job of saving the binary and also updating the everything
						gnvm.PostBinaryDownload(nodejs_version)
						break
					}
				}
			}
		}
		return nil
	})
}

func (gnvm *GNVM) PostBinaryDownload(nodejs_version NodeJS) error {
	// extract save the paths correctly then do this
	// unzip the file and do the rest

	// return gnvm.Db.Update(func(t *bolt.Tx) error {
	// 	installed_node_versions, err := t.CreateBucketIfNotExists([]byte(INSTALLED_BUCKET))

	// 	if err != nil {
	// 		return err
	// 	}

	// 	buf, err := json.Marshal(nodejs_version)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	return installed_node_versions.Put(
	// 		[]byte(nodejs_version.Version),
	// 		buf,
	// 	)
	// })
	return nil
}

func (gnvm *GNVM) UninstallNodeJSVersion(version string) error {
	return gnvm.Db.Update(func(t *bolt.Tx) error {
		// first get the key get the path uninstall everything from the path then delete the key from the db
		return t.Bucket([]byte(INSTALLED_BUCKET)).Delete([]byte(version))
	})
}

func (gnvm *GNVM) UseNodeJSVersion(version string) error {
	// Retrieve the key again.
	return gnvm.Db.View(func(tx *bolt.Tx) error {
		value := tx.Bucket([]byte(INSTALLED_BUCKET)).Get([]byte(version))
		// we get the installed version suggested and then do stuff with it
		if value != nil {
			var nodejs_version NodeJS

			if err := json.Unmarshal(value, &nodejs_version); err != nil {
				return err
			}

			// use the value and set the path of the mnode.cmd to this
			fmt.Println("We are switching to this node version ", nodejs_version.Version)
		}
		return nil
	})
}
