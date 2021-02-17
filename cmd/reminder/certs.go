package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type (
	certItem struct {
		Name  string
		Error error
		Cert  *x509.Certificate
	}

	ByCertNotAfterAsc []certItem
)

func (a ByCertNotAfterAsc) Len() int           { return len(a) }
func (a ByCertNotAfterAsc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCertNotAfterAsc) Less(i, j int) bool { return a[i].Cert.NotAfter.Before(a[j].Cert.NotAfter) }

func getCertsInfo() []item {
	result, err := oc.List(certsDir, false)
	if err != nil {
		return []item{
			{
				Name:  "Error",
				Date:  err.Error(),
				Class: "text-danger",
			},
		}
	}
	files := []string{}
	for _, f := range result.Files {
		if !strings.HasSuffix(f.Name, ".cert") {
			continue
		}
		files = append(files, f.Name)
	}
	itemsChan := make(chan certItem, len(files))
	var wg sync.WaitGroup
	wg.Add(len(files))
	for _, f := range files {
		go func(f string) {
			fn := strings.TrimPrefix(f, certsDir)
			err := func() error {
				var buffer bytes.Buffer
				_, err := oc.Download(f, &buffer)
				if err != nil {
					return err
				}
				content, err := decrypt(buffer.Bytes())
				if err != nil {
					return err
				}
				block, _ := pem.Decode(content)
				cert, err := x509.ParseCertificate(block.Bytes)
				if err != nil {
					return err
				}
				itemsChan <- certItem{
					Name: fn,
					Cert: cert,
				}
				return nil
			}()
			if err != nil {
				itemsChan <- certItem{
					Name:  fn,
					Error: err,
				}
			}
			wg.Done()
		}(f)
	}
	wg.Wait()
	close(itemsChan)
	certs := []certItem{}
	for i := range itemsChan {
		certs = append(certs, i)
	}
	sort.Sort(ByCertNotAfterAsc(certs))
	items := []item{}
	for _, c := range certs {
		if c.Error == nil {
			days := int(time.Until(c.Cert.NotAfter).Hours() / 24)
			class := "text-danger"
			if days > 30 {
				class = "text-success"
			} else if days > 14 {
				class = "text-warning"
			}
			items = append(items, item{
				Name:  c.Name,
				Date:  fmt.Sprintf("%s (%d days)", c.Cert.NotAfter.Format("2006-01-02"), days),
				Class: class,
			})
		} else {
			items = append(items, item{
				Name:  c.Name,
				Date:  c.Error.Error(),
				Class: "text-danger",
			})
		}
	}
	return items
}

func decrypt(content []byte) ([]byte, error) {
	block, err := aes.NewCipher([]byte(certsEncryptionKey))
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesgcm.NonceSize()
	nonce, ciphertext := content[:nonceSize], content[nonceSize:]

	return aesgcm.Open(nil, nonce, ciphertext, nil)
}
