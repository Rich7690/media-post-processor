package controllers

type Config struct {
	RadarrApiKey         string `json:"RadarrApiKey"`
	RadarrEndpoint       string `json:"RadarrEndpoint"`
	SonarrApiKey         string `json:"SonarrApiKey"`
	SonarrEndpoint       string `json:"SonarrEndpoint"`
	WorkerEnabled        bool   `json:"WorkerEnabled"`
	RadarrScannerEnabled bool   `json:"RadarrScannerEnabled"`
	SonarrScannerEnabled bool   `json:"SonarrScannerEnabled"`
}

/*
func SecretKey(key string) string {

	keyLen := len(key)
	if keyLen == 0 || keyLen == 1 {
		return key
	}

	amountToTrim := int(math.Ceil(.5 * float64(keyLen)))

	secret := strings.Repeat("*", amountToTrim)

	return key[0:(len(key)-amountToTrim)] + secret
}

func SecretUrl(urlString string) string {
	u, err := url.Parse(urlString)
	if err != nil {
		return ""
	}
	if u == nil {
		return ""
	}

	if u.User != nil {

		_, set := u.User.Password()
		if set {
			secretUser := url.UserPassword(u.User.Username(), strings.Repeat("$", 3))
			u.User = secretUser
			return u.String()
		} else {
			secretUser := url.User(strings.Repeat("$", 3))
			u.User = secretUser
			return u.String()
		}


	}
	return u.String()
}

func GetConfigHandler(c *gin.Context) {

	conf := Config{
	}

	c.JSON(200, conf)
}*/
