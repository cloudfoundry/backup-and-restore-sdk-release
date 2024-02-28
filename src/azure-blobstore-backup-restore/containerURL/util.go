package containerURL

import "net/url"

func Sanitise(urlToSanitise string) string {
	parsedURL, err := url.Parse(urlToSanitise)
	if err != nil {
		// This was written on 2024-02-26 for use in logging. If it
		// gets an invalid URL, we want to log that invalid URL for
		// debugging.
		return urlToSanitise
	}
	queryParams := parsedURL.Query()
	for key := range queryParams {
		if key != "snapshot" {
			queryParams.Set(key, "REDACTED")
		}
	}

	parsedURL.RawQuery = queryParams.Encode()
	return parsedURL.String()
}
