package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/slack-go/slack"
)

const (
	listenAddr  = "localhost:9876"
	redirectURI = "https://localhost:9876/callback"
	userScopes  = "reactions:write,channels:history,channels:read"
	timeout     = 2 * time.Minute
)

// Run は Slack OAuth 2.0 フローを実行し、User Token (xoxp-...) を返す。
func Run(ctx context.Context, clientID, clientSecret string) (string, error) {
	state, err := randomState()
	if err != nil {
		return "", fmt.Errorf("state生成に失敗: %w", err)
	}

	// コールバック結果を受け取るチャネル
	type result struct {
		code string
		err  error
	}
	ch := make(chan result, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		if errParam := q.Get("error"); errParam != "" {
			ch <- result{err: fmt.Errorf("slack認可エラー: %s – %s", errParam, q.Get("error_description"))}
			fmt.Fprintln(w, "認可に失敗しました。ターミナルを確認してください。")
			return
		}

		if q.Get("state") != state {
			ch <- result{err: fmt.Errorf("stateが一致しません（CSRF検証失敗）")}
			fmt.Fprintln(w, "state不一致エラー。ターミナルを確認してください。")
			return
		}

		ch <- result{code: q.Get("code")}
		fmt.Fprintln(w, "認可が完了しました！このタブは閉じてOKです。")
	})

	// 自己署名証明書を生成してTLSリスナーを作成
	tlsCert, err := generateSelfSignedCert()
	if err != nil {
		return "", fmt.Errorf("TLS証明書の生成に失敗: %w", err)
	}

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return "", fmt.Errorf("ローカルサーバー起動失敗 (%s): %w", listenAddr, err)
	}
	defer ln.Close()

	tlsLn := tls.NewListener(ln, &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
	})

	srv := &http.Server{Handler: mux}
	go srv.Serve(tlsLn) //nolint:errcheck
	defer srv.Close()

	// 認可URLを構築してブラウザで開く
	authURL := fmt.Sprintf(
		"https://slack.com/oauth/v2/authorize?client_id=%s&user_scope=%s&redirect_uri=%s&state=%s",
		clientID, userScopes, redirectURI, state,
	)

	fmt.Println("ブラウザで Slack 認可ページを開きます...")
	fmt.Println("※ コールバック時にブラウザが証明書の警告を出す場合があります。")
	fmt.Println("  「詳細設定」→「localhostにアクセスする」で続行してください。")
	fmt.Printf("\n自動で開かない場合は以下のURLをコピーしてください:\n%s\n\n", authURL)
	openBrowser(authURL)

	// コールバック待ち（タイムアウト付き）
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var code string
	select {
	case res := <-ch:
		if res.err != nil {
			return "", res.err
		}
		code = res.code
	case <-ctx.Done():
		return "", fmt.Errorf("タイムアウト（%s以内に認可が完了しませんでした）", timeout)
	}

	// コード → トークン交換
	resp, err := slack.GetOAuthV2Response(http.DefaultClient, clientID, clientSecret, code, redirectURI)
	if err != nil {
		return "", fmt.Errorf("トークン交換に失敗: %w", err)
	}

	token := resp.AuthedUser.AccessToken
	if token == "" {
		return "", fmt.Errorf("User Tokenが取得できませんでした（user_scopeの設定を確認してください）")
	}
	return token, nil
}

// generateSelfSignedCert はlocalhostの自己署名証明書をメモリ上に生成する。
func generateSelfSignedCert() (tls.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return tls.Certificate{}, err
	}

	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(1 * time.Hour), // 1時間だけ有効
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}

	return tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  key,
	}, nil
}

// randomState はCSRF防止用のランダム文字列を生成する。
func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// openBrowser はOSに応じてデフォルトブラウザでURLを開く。
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return
	}
	_ = cmd.Start()
}
