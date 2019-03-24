package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"

	"github.com/go-ini/ini"
	input "github.com/tcnksm/go-input"
	"golang.org/x/oauth2"
	resty "gopkg.in/resty.v1"
)

func main() {
	cfg, err := ini.Load("config.ini")
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}

	ui := &input.UI{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}

	fmt.Println(os.Args[1])

	query := "Is that your key ? [Y/n]"
	res, err := ui.Ask(query, &input.Options{
		Default:     "Y",
		Required:    false,
		HideOrder:   true,
		HideDefault: true,
		Loop:        true,
		ValidateFunc: func(s string) error {
			matched, err := regexp.MatchString("^[YyNn]\\w+|^[YyNn]", s)
			if err != nil {
				log.Fatal(err)
			}
			if matched {
				return nil
			} else {
				return fmt.Errorf("input must be Y or N")
			}
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	matched, err := regexp.MatchString("^[Yy]\\w+|^[Yy]", res)
	if matched {
		fmt.Println("Okay! Welcome to Akmey")
	} else {
		fmt.Println("Try to reconfigure your SSH client to use the key you want to manage Akmey")
	}

	for {
		fmt.Println()
		choice, err := ui.Select("What do you want to do ?", []string{"Add my key to Akmey", "Remove my key", "Edit my key", "About Akmey", "Exit"}, &input.Options{
			Default: "Add my key to Akmey",
			Loop:    true,
		})
		if err != nil {
			log.Fatal(err)
		}
		switch choice {
		case "Add my key to Akmey":
			fmt.Println("Login to your Akmey account to add your key, if you don't have an Akmey account, please register here : " + cfg.Section("clientlink").Key("registerurl").String())
			email, err := ui.Ask("E-Mail", &input.Options{
				HideOrder: true,
				Required:  true,
				Loop:      true,
			})
			if err != nil {
				log.Fatal(err)
			}
			password, err := ui.Ask("Password", &input.Options{
				HideOrder:   true,
				Loop:        true,
				Required:    true,
				Mask:        true,
				MaskDefault: true,
			})
			if err != nil {
				log.Fatal(err)
			}
			ctx := context.Background()
			conf := &oauth2.Config{
				ClientID:     cfg.Section("clientlink").Key("clientid").String(),
				ClientSecret: cfg.Section("clientlink").Key("clientsecret").String(),
				Scopes:       []string{"keys"},
				Endpoint: oauth2.Endpoint{
					AuthURL:  "http://no.where",
					TokenURL: cfg.Section("clientlink").Key("url").String() + "/oauth/token",
				},
			}
			token, err := conf.PasswordCredentialsToken(ctx, email, password)
			if err != nil {
				fmt.Println("Incorrect credentials")
				continue
			}
			resp, err := resty.R().
				SetHeader("Authorization", "Bearer "+token.AccessToken).
				SetHeader("Content-Type", "application/x-www-form-urlencoded").
				SetHeader("Accept", "application/json").
				SetFormData(map[string]string{
					"key": os.Args[1],
				}).Post(cfg.Section("clientlink").Key("url").String() + "/api/keys/add")
			if err != nil {
				log.Fatal(err)
			}
			var f interface{}
			err = json.Unmarshal(resp.Body(), &f)
			if err != nil {
				log.Fatal(err)
			}
			parsed := f.(map[string]interface{})
			if parsed["success"] != true {
				fmt.Println("The server refused our key, the key is already used or you are not allowed to add keys.")
				continue
			} else {
				fmt.Println("Great ! Your key is registered on Akmey.")
				continue
			}
		case "Remove my key":
			fmt.Println("Login to your Akmey account to remove your key, if you don't have an Akmey account, please register here : " + cfg.Section("clientlink").Key("registerurl").String())
			email, err := ui.Ask("E-Mail", &input.Options{
				HideOrder: true,
				Required:  true,
				Loop:      true,
			})
			if err != nil {
				log.Fatal(err)
			}
			password, err := ui.Ask("Password", &input.Options{
				HideOrder:   true,
				Loop:        true,
				Required:    true,
				Mask:        true,
				MaskDefault: true,
			})
			if err != nil {
				log.Fatal(err)
			}
			ctx := context.Background()
			conf := &oauth2.Config{
				ClientID:     cfg.Section("clientlink").Key("clientid").String(),
				ClientSecret: cfg.Section("clientlink").Key("clientsecret").String(),
				Scopes:       []string{"keys"},
				Endpoint: oauth2.Endpoint{
					AuthURL:  "http://no.where",
					TokenURL: cfg.Section("clientlink").Key("url").String() + "/oauth/token",
				},
			}
			token, err := conf.PasswordCredentialsToken(ctx, email, password)
			if err != nil {
				fmt.Println("Incorrect credentials")
				continue
			}
			resp, err := resty.R().
				SetHeader("Authorization", "Bearer "+token.AccessToken).
				SetHeader("Content-Type", "application/x-www-form-urlencoded").
				SetHeader("Accept", "application/json").
				SetFormData(map[string]string{
					"key": os.Args[1],
				}).Post(cfg.Section("clientlink").Key("url").String() + "/api/keys/fetch")
			if err != nil {
				log.Fatal(err)
			}
			var f interface{}
			err = json.Unmarshal(resp.Body(), &f)
			if err != nil {
				log.Fatal(err)
			}
			parsed := f.(map[string]interface{})
			if parsed["id"] != 0 {
				resp, err := resty.R().
					SetHeader("Authorization", "Bearer "+token.AccessToken).
					SetHeader("Content-Type", "application/x-www-form-urlencoded").
					SetHeader("Accept", "application/json").
					Delete(cfg.Section("clientlink").Key("url").String() + "/api/keys/" + strconv.FormatFloat(parsed["id"].(float64), 'f', 0, 64))
				if err != nil {
					log.Fatal(err)
				}
				var f interface{}
				err = json.Unmarshal(resp.Body(), &f)
				if err != nil {
					log.Fatal(err)
				}
				parsed := f.(map[string]interface{})
				if parsed["success"] != true {
					fmt.Println(parsed["message"])
				} else {
					fmt.Println("The key is now out of our servers, goodbye!")
				}
			} else {
				fmt.Println(parsed["message"])
			}
		case "Edit my key":
			fmt.Println("Login to your Akmey account to edit your key, if you don't have an Akmey account, please register here : " + cfg.Section("client-link").Key("registerurl").String())
			email, err := ui.Ask("E-Mail", &input.Options{
				HideOrder: true,
				Required:  true,
				Loop:      true,
			})
			if err != nil {
				log.Fatal(err)
			}
			password, err := ui.Ask("Password", &input.Options{
				HideOrder:   true,
				Loop:        true,
				Required:    true,
				Mask:        true,
				MaskDefault: true,
			})
			if err != nil {
				log.Fatal(err)
			}
			ctx := context.Background()
			conf := &oauth2.Config{
				ClientID:     cfg.Section("client-link").Key("clientid").String(),
				ClientSecret: cfg.Section("client-link").Key("clientsecret").String(),
				Scopes:       []string{"keys"},
				Endpoint: oauth2.Endpoint{
					AuthURL:  "http://no.where",
					TokenURL: cfg.Section("client-link").Key("url").String() + "/oauth/token",
				},
			}
			token, err := conf.PasswordCredentialsToken(ctx, email, password)
			if err != nil {
				fmt.Println("Incorrect credentials")
				continue
			}
			resp, err := resty.R().
				SetHeader("Authorization", "Bearer "+token.AccessToken).
				SetHeader("Content-Type", "application/x-www-form-urlencoded").
				SetHeader("Accept", "application/json").
				SetFormData(map[string]string{
					"key": os.Args[1],
				}).Post(cfg.Section("clientlink").Key("url").String() + "/api/keys/fetch")
			if err != nil {
				log.Fatal(err)
			}
			var f interface{}
			err = json.Unmarshal(resp.Body(), &f)
			if err != nil {
				log.Fatal(err)
			}
			parsed := f.(map[string]interface{})
			if parsed["id"] != 0 {
				comment, err := ui.Ask("Insert here your new key comment (key name)", &input.Options{
					HideOrder: true,
					Required:  true,
					Loop:      true,
				})
				if err != nil {
					log.Fatal(err)
				}
				resp, err := resty.R().
					SetHeader("Authorization", "Bearer "+token.AccessToken).
					SetHeader("Content-Type", "application/x-www-form-urlencoded").
					SetHeader("Accept", "application/json").
					SetFormData(map[string]string{
						"comment": comment,
					}).
					Put(cfg.Section("clientlink").Key("url").String() + "/api/keys/" + strconv.FormatFloat(parsed["id"].(float64), 'f', 0, 64))
				if err != nil {
					log.Fatal(err)
				}
				var f interface{}
				err = json.Unmarshal(resp.Body(), &f)
				if err != nil {
					log.Fatal(err)
				}
				parsed := f.(map[string]interface{})
				if parsed["success"] != true {
					fmt.Println(parsed["message"])
				} else {
					fmt.Println("Your key is named " + comment + " now.")
				}
			} else {
				fmt.Println(parsed["message"])
			}
		case "About Akmey":
			fmt.Println("Akmey is a SSH keyserver, it is shipped with tools to add user keys to your authorized_keys files automatically.")
		case "Exit":
			os.Exit(0)
		}
	}

}
