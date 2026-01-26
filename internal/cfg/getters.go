package cfg

import (
	"fmt"
	"regexp"
	"time"
)

func UserAgent() string {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.userAgent
}

func ListenAddr() string {
	lock.RLock()
	defer lock.RUnlock()
	return fmt.Sprintf("%s:%d", globalConfig.bindAddr, globalConfig.listenPort)
}

func ConfigDir() string {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.configDir
}

func DatabaseUrl() string {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.databaseUrl
}

func MusicBrainzUrl() string {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.musicBrainzUrl
}

func MusicBrainzRateLimit() int {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.musicBrainzRateLimit
}

func LogLevel() int {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.logLevel
}

func StructuredLogging() bool {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.structuredLogging
}

func LbzRelayEnabled() bool {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.lbzRelayEnabled
}

func LbzRelayUrl() string {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.lbzRelayUrl
}

func LbzRelayToken() string {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.lbzRelayToken
}

func DefaultPassword() string {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.defaultPw
}

func DefaultUsername() string {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.defaultUsername
}

func DefaultTheme() string {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.defaultTheme
}

func FullImageCacheEnabled() bool {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.enableFullImageCache
}

func DeezerDisabled() bool {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.disableDeezer
}

func CoverArtArchiveDisabled() bool {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.disableCAA
}

func MusicBrainzDisabled() bool {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.disableMusicBrainz
}

func SubsonicEnabled() bool {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.subsonicEnabled
}

func SubsonicUrl() string {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.subsonicUrl
}

func SubsonicParams() string {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.subsonicParams
}

func LastFMApiKey() string {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.lastfmApiKey
}

func SkipImport() bool {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.skipImport
}

func AllowedHosts() []string {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.allowedHosts
}

func AllowAllHosts() bool {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.allowAllHosts
}

func AllowedOrigins() []string {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.allowedOrigins
}

func RateLimitDisabled() bool {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.disableRateLimit
}

func ThrottleImportMs() int {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.importThrottleMs
}

// returns the before, after times, in that order
func ImportWindow() (time.Time, time.Time) {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.importBefore, globalConfig.importAfter
}

func FetchImagesDuringImport() bool {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.fetchImageDuringImport
}

func ArtistSeparators() []*regexp.Regexp {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.artistSeparators
}

func LoginGate() bool {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.loginGate
}

func ForceTZ() *time.Location {
	lock.RLock()
	defer lock.RUnlock()
	return globalConfig.forceTZ
}
