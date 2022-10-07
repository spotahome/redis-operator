package util

import "regexp"

func GetHashedPasswordsFromRedisUserConf(userConf string) []string {
	re := regexp.MustCompile(`#[a-zA-Z0-1]+`)
	passwords := re.FindAllString(userConf, -1)
	return passwords
}

// Returns username from a string containing userspec - in the format returned by `acl list`
func GetUsernameFromRedisUserConf(userConf string) string {
	re := regexp.MustCompile("user ([a-z0-9A-Z-]+)")
	matches := re.FindStringSubmatch(userConf)
	if nil != matches {
		return string(matches[1])

	}
	return ""
}
