package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"strings"
	"syscall"
	"time"

	"github.com/jcmturner/passvault/entry"
	"github.com/jcmturner/passvault/store"
	term "golang.org/x/crypto/ssh/terminal"

	"github.com/atotto/clipboard"
	"github.com/prologic/bitcask"
)

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		st, err := term.GetState(int(syscall.Stdin))
		if err == nil {
			for sig := range c {
				fmt.Fprintf(os.Stdin, "%s Exited...\n", sig.String())
				term.Restore(int(syscall.Stdin), st)
				os.Exit(0)
			}
		}
	}()

	usr, err := user.Current()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting user's details: %v\n", err)
		os.Exit(1)
	}
	vault := fmt.Sprintf("%s/passvault", usr.HomeDir)

	name := flag.String("name", "", "Name of the secret entry")
	add := flag.Bool("add", false, "Add entry")
	prnt := flag.Bool("print", false, "Print details to stdout")
	json := flag.Bool("json", false, "Output in JSON format")
	sess := flag.Bool("session", false, "Run as non-exiting read only session")
	list := flag.Bool("list", false, "List the names of all the secret entries")
	del := flag.Bool("del", false, "Delete a secret entry")
	vpath := flag.String("vault", vault, "Specify path to the vault directory")
	flag.Parse()

	cask, err := bitcask.Open(*vpath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error openning vault: %v\n", err)
		os.Exit(1)
	}
	defer cask.Close()

	if *list {
		fmt.Println("Entries List:")
		err := cask.Fold(func(k string) error {
			fmt.Println(k)
			return nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing entries: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if *del {
		if *name == "" {
			*name = entryNamePrompt()
		}
		fmt.Fprintf(os.Stderr, "Confirm deletion of secret entry %s with passphrase\n", *name)
		passphrase := passwordPrompt()
		_, err := store.Get(*name, passphrase, cask)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to validate entry for deletion: %v\n", err)
			os.Exit(1)
		}
		err = cask.Delete(*name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to delete %s: %v\n", *name, err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Entry %s deleted successfully\n", *name)
		os.Exit(0)
	}

	passphrase := passwordPrompt()
	t := time.Now()

	if *sess {
		for {
			fmt.Println("------")
			*name = entryNamePrompt()
			if time.Now().After(t.Add(time.Hour * 8)) {
				passphrase = passwordPrompt()
				t = time.Now()
			}
			err = get(*name, passphrase, cask, false, false)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting entry: %v\n", err)
			}
		}
	} else {
		if *name == "" {
			*name = entryNamePrompt()
		}

		var e entry.Entry
		if *add {
			e.Name = *name
			e.Secret.Username, e.Secret.Password = newEntryDetails()
			err = store.Put(e.Name, e.Secret.Username, e.Secret.Password, passphrase, cask)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error storing secret: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Secret stored successfully")
			os.Exit(0)
		}
		err = get(*name, passphrase, cask, *prnt, *json)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting entry: %v\n", err)
			os.Exit(1)
		}
	}
}

func passwordPrompt() []byte {
	var p []byte
	for len(p) < 1 {
		fmt.Fprint(os.Stderr, "Enter passphrase: ")
		p, _ = term.ReadPassword(int(syscall.Stdin))
		fmt.Println("")
	}
	return p
}

func entryNamePrompt() string {
	var n string
	reader := bufio.NewReader(os.Stdin)
	for n == "" {
		fmt.Fprint(os.Stderr, "Enter name of secret entry: ")
		n, _ = reader.ReadString('\n')
		n = strings.TrimSpace(n)
	}
	return n
}

func newEntryDetails() (string, string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Fprint(os.Stderr, "Username: ")
	username, _ := reader.ReadString('\n')
	fmt.Fprint(os.Stderr, "Password: ")
	password, _ := term.ReadPassword(int(syscall.Stdin))
	fmt.Println("")
	return strings.TrimSpace(username), strings.TrimSpace(string(password))
}

func get(name string, passphrase []byte, cask *bitcask.Bitcask, print, json bool) error {
	e, err := store.Get(name, passphrase, cask)
	if err != nil {
		return fmt.Errorf("error getting secret %s: %v", name, err)
	}
	if print {
		if json {
			fmt.Println(e.JSON())
		} else {
			fmt.Println(e.String())
		}
	} else {
		fmt.Println(e.Redacted())
		err = clipboard.WriteAll(e.Secret.Password)
		if err != nil {
			return fmt.Errorf("error writing secret's password to clipboard: %v", err)
		}
		fmt.Fprintln(os.Stderr, "Password copied to clipboard.")
	}
	return nil
}
