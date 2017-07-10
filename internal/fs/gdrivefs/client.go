package gdrivefs

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	drive "google.golang.org/api/drive/v3"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func (d *GDriveFS) getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := d.tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}

	tok, err := d.tokenFromFile(cacheFile)
	if err != nil {
		tok = d.getTokenFromWeb(config)
		d.saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)

}

func (d *GDriveFS) tokenCacheFile() (string, error) {
	tokenCacheDir := d.core.Config.HomeDir

	err := os.MkdirAll(tokenCacheDir, 0700)

	return filepath.Join(tokenCacheDir, url.QueryEscape("auth.json")), err

}

func (d *GDriveFS) getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	fmt.Printf(
		`Go to the following link in your browser: 
----------------------------------------------------------------------------------------------
	%v
----------------------------------------------------------------------------------------------

type the authorization code: `, authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}

	return tok
}

func (d *GDriveFS) tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	return t, err
}

func (d *GDriveFS) saveToken(file string, token *oauth2.Token) {
	log.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v\n", err)
	}
	defer f.Close()

	json.NewEncoder(f).Encode(token)
}

/*func getConfigPath() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(usr.HomeDir, ".gdrivemount")

	return configDir, nil
}*/

// Init driveService
func (d *GDriveFS) initClient() {

	configPath := d.core.Config.HomeDir

	ctx := context.Background() // Context from GDriveFS

	b, err := ioutil.ReadFile(filepath.Join(configPath, "client_secret.json"))
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file: %v", err)
	}

	client := d.getClient(ctx, config)
	d.client, err = drive.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve drive Client: %v", err)
	}

}