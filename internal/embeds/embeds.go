package embeds

import "strings"

const (

	// Domains to replace
	xDomain         = "https://x.com"
	twitterDomain   = "https://twitter.com"
	facebookDomain  = "https://facebook.com"
	instagramDomain = "https://instagram.com"

	// Embed solutions - in the future we'll do our own parsing but these work
	vxTwitterDomain = "https://vxtwitter.com"
	facebedDomain   = "https://facebed.com"
	instaBedDomain  = "https://eeinstagram.com"
)

var instagramPostTypes = []string{"/p/", "/reel/", "/reels/", "/tv/", "/stories/"}

func handleMessage(message string) {
	if strings.Contains(xDomain, message) {
		message = strings.ReplaceAll(message, xDomain, vxTwitterDomain)
	}

	if strings.Contains(twitterDomain, message) {
		message = strings.ReplaceAll(message, twitterDomain, vxTwitterDomain)
	}

	if strings.Contains(facebookDomain, message) {
		message = strings.ReplaceAll(message, facebookDomain, facebedDomain)
	}

	if strings.Contains(instagramDomain, message) {
		for _, postType := range instagramPostTypes {
			if strings.Contains(message, postType) {
				message = strings.ReplaceAll(message, instagramDomain, instaBedDomain)
				break
			}
		}
	}
}
