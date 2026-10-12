// Command coachctl creates coach accounts and resets their passwords, since the API has no public signup.
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/AbolfazlTayyar/gogym/internal/coach"
	"github.com/AbolfazlTayyar/gogym/internal/platform/config"
	"github.com/AbolfazlTayyar/gogym/internal/platform/db"
)

const usage = `usage:
  coachctl create -first-name NAME -last-name NAME -phone 09XXXXXXXXX
  coachctl reset-password -phone 09XXXXXXXXX

The password is read from the first line of stdin, so it stays out of shell history
and the process list.`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "coachctl:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New(usage)
	}

	switch args[0] {
	case "create":
		return create(args[1:])
	case "reset-password":
		return resetPassword(args[1:])
	default:
		return fmt.Errorf("unknown command %q\n%s", args[0], usage)
	}
}

func create(args []string) error {
	flags := flag.NewFlagSet("create", flag.ContinueOnError)
	firstName := flags.String("first-name", "", "coach's first name")
	lastName := flags.String("last-name", "", "coach's last name")
	phone := flags.String("phone", "", "login phone number")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *firstName == "" || *lastName == "" || *phone == "" {
		return errors.New("create needs -first-name, -last-name and -phone")
	}

	password, err := readPassword(os.Stdin)
	if err != nil {
		return err
	}

	svc, err := newService()
	if err != nil {
		return err
	}

	created, err := svc.Create(context.Background(), coach.CreateInput{
		FirstName: *firstName,
		LastName:  *lastName,
		Phone:     *phone,
		Password:  password,
	})
	if err != nil {
		return err
	}

	fmt.Printf("created coach %s (%s)\n", created.ID, created.Phone)
	return nil
}

func resetPassword(args []string) error {
	flags := flag.NewFlagSet("reset-password", flag.ContinueOnError)
	phone := flags.String("phone", "", "login phone number of the coach")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *phone == "" {
		return errors.New("reset-password needs -phone")
	}

	password, err := readPassword(os.Stdin)
	if err != nil {
		return err
	}

	svc, err := newService()
	if err != nil {
		return err
	}

	if err := svc.ResetPassword(context.Background(), *phone, password); err != nil {
		return err
	}

	fmt.Printf("password reset for %s; tokens issued before now stay valid until they expire\n", *phone)
	return nil
}

func newService() (*coach.Service, error) {
	cfg := config.Load()

	gormDB, err := db.New(cfg)
	if err != nil {
		return nil, err
	}

	return coach.NewService(coach.NewRepository(gormDB), cfg.JWTSecret, cfg.JWTExpiry), nil
}

// readPassword prompts only on a terminal, and on stderr, so piped use leaves stdout to the result.
func readPassword(in *os.File) (string, error) {
	if info, err := in.Stat(); err == nil && info.Mode()&os.ModeCharDevice != 0 {
		fmt.Fprint(os.Stderr, "password: ")
	}

	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("failed to read password: %w", err)
	}

	password := strings.TrimRight(line, "\r\n")
	if password == "" {
		return "", errors.New("no password on stdin")
	}
	return password, nil
}
