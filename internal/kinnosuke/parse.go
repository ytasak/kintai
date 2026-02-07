package kinnosuke

import (
	"errors"
	"fmt"
	"os"
	"regexp"
)

var (
	reAuthorized = regexp.MustCompile(`<div class="user_name">`)
	reCSRF       = regexp.MustCompile(`name="(__sectag_[0-9a-f]+)" value="([0-9a-f]+)"`)
	reStartTime  = regexp.MustCompile(`>出社<br(?:\s*\/)?>\((\d\d:\d\d)\)`)
	reLeaveTime  = regexp.MustCompile(`>退社<br(?:\s*\/)?>\((\d\d:\d\d)\)`)
)

type credential struct {
	CompanyCD string
	LoginCD   string
	Password  string
}

func loadCredentialFromEnv() (credential, error) {
	c := credential{
		CompanyCD: os.Getenv("KIN_COMPANYCD"),
		LoginCD:   os.Getenv("KIN_LOGINCD"),
		Password:  os.Getenv("KIN_PASSWORD"),
	}
	if c.CompanyCD == "" || c.LoginCD == "" || c.Password == "" {
		return credential{}, errors.New("missing env: KIN_COMPANYCD / KIN_LOGINCD / KIN_PASSWORD")
	}
	return c, nil
}

func authorized(html string) bool { return reAuthorized.MatchString(html) }

func csrfToken(html string) (key, value string, ok bool) {
	m := reCSRF.FindStringSubmatch(html)
	if m == nil || len(m) < 3 {
		return "", "", false
	}
	return m[1], m[2], true
}

func startTime(html string) (string, bool) {
	m := reStartTime.FindStringSubmatch(html)
	if m == nil || len(m) < 2 {
		return "", false
	}
	return m[1], true
}

func leaveTime(html string) (string, bool) {
	m := reLeaveTime.FindStringSubmatch(html)
	if m == nil || len(m) < 2 {
		return "", false
	}
	return m[1], true
}

func login(cli *Client, cred credential) error {
	_, err := cli.PostForm(map[string]string{
		"module":      "login",
		"y_companycd": cred.CompanyCD,
		"y_logincd":   cred.LoginCD,
		"password":    cred.Password,
		"trycnt":      "1",
	})
	return err
}

func stamp(cli *Client, stampingType string, tokenKey string, tokenVal string) error {
	_, err := cli.PostForm(map[string]string{
		"module":                     "timerecorder",
		"action":                     "timerecorder",
		tokenKey:                     tokenVal,
		"timerecorder_stamping_type": stampingType,
	})
	return err
}

// 拡張より簡略：毎回ログイン前提でもOKだが、authorizedならスキップする
func ensureAuthorized(cli *Client, cred credential) (string, error) {
	top, err := cli.GetTopHTML()
	if err != nil {
		return "", err
	}
	if authorized(top) {
		return top, nil
	}
	if err := login(cli, cred); err != nil {
		return "", fmt.Errorf("login failed: %w", err)
	}
	top, err = cli.GetTopHTML()
	if err != nil {
		return "", err
	}
	if !authorized(top) {
		return "", errors.New("still unauthorized after login (credentials or SSO issue)")
	}
	return top, nil
}

func StampStart() (string, error) { return doStamp("1") } // 出社
func StampEnd() (string, error)   { return doStamp("2") } // 退社

func doStamp(stType string) (string, error) {
	cred, err := loadCredentialFromEnv()
	if err != nil {
		return "", err
	}
	cli, err := New()
	if err != nil {
		return "", err
	}

	top, err := ensureAuthorized(cli, cred)
	if err != nil {
		return "", err
	}

	tk, tv, ok := csrfToken(top)
	if !ok {
		return "", errors.New("csrf token not found in top html")
	}

	if err := stamp(cli, stType, tk, tv); err != nil {
		return "", fmt.Errorf("stamp failed: %w", err)
	}

	after, err := cli.GetTopHTML()
	if err != nil {
		return "", err
	}

	if stType == "1" {
		if t, ok := startTime(after); ok {
			return t, nil
		}
		return "", errors.New("stamp may have failed: start time not found after stamping")
	}

	if t, ok := leaveTime(after); ok {
		return t, nil
	}
	return "", errors.New("stamp may have failed: leave time not found after stamping")
}
