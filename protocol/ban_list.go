package protocol

import (
	"bufio"
	"fnd.localhost/handshake/primitives"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

const (
	CurrentBanListVersion = 1
)

var verRegex = regexp.MustCompile("^v([\\d]+)$")

func ParseBanListVersion(line string) (int, error) {
	splits := strings.Split(line, ":")
	if len(splits) != 2 {
		return 0, errors.New("ban list version must consist of two colon-separated components")
	}
	if splits[0] != "FNBAN" {
		return 0, errors.New("ban list version must start with FNBAN")
	}
	if !verRegex.MatchString(splits[1]) {
		return 0, errors.New("ban list version must end with v followed by a digit")
	}
	verStr := strings.TrimPrefix(splits[1], "v")
	verInt, err := strconv.Atoi(verStr)
	if err != nil {
		// should not happen given
		// regex check above
		panic(err)
	}
	return verInt, nil
}

func ReadBanList(r io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(r)
	if !scanner.Scan() {
		return nil, errors.New("ban list must start with version line")
	}
	firstLine := scanner.Text()
	version, err := ParseBanListVersion(firstLine)
	if err != nil {
		return nil, err
	}
	if version != CurrentBanListVersion {
		return nil, errors.New("unsupported ban list version")
	}

	var names []string
	i := 1
	for scanner.Scan() {
		name := strings.Trim(scanner.Text(), " \t")
		if err := primitives.ValidateName(name); err != nil {
			return nil, errors.Wrapf(err, "invalid name on line %d: %s", i, name)
		}
		names = append(names, name)
		i++
	}
	return names, nil
}

func FetchListFile(url string) ([]string, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch list")
	}
	defer res.Body.Close()
	return ReadBanList(res.Body)
}
