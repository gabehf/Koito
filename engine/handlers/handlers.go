// package handlers implements route handlers
package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
)

const defaultLimitSize = 100
const maximumLimit = 500

func OptsFromRequest(r *http.Request) db.GetItemsOpts {
	l := logger.FromContext(r.Context())

	l.Debug().Msg("OptsFromRequest: Parsing query parameters")

	limitStr := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		l.Debug().Msgf("OptsFromRequest: Query parameter 'limit' not specified, using default %d", defaultLimitSize)
		limit = defaultLimitSize
	}
	if limit > maximumLimit {
		l.Debug().Msgf("OptsFromRequest: Limit exceeds maximum %d, using default %d", maximumLimit, defaultLimitSize)
		limit = defaultLimitSize
	}

	pageStr := r.URL.Query().Get("page")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		l.Debug().Msg("OptsFromRequest: Page parameter is less than 1, defaulting to 1")
		page = 1
	}

	artistIdStr := r.URL.Query().Get("artist_id")
	artistId, _ := strconv.Atoi(artistIdStr)
	albumIdStr := r.URL.Query().Get("album_id")
	albumId, _ := strconv.Atoi(albumIdStr)
	trackIdStr := r.URL.Query().Get("track_id")
	trackId, _ := strconv.Atoi(trackIdStr)

	tf := TimeframeFromRequest(r)

	var period db.Period
	switch strings.ToLower(r.URL.Query().Get("period")) {
	case "day":
		period = db.PeriodDay
	case "week":
		period = db.PeriodWeek
	case "month":
		period = db.PeriodMonth
	case "year":
		period = db.PeriodYear
	case "all_time":
		period = db.PeriodAllTime
	}

	l.Debug().Msgf("OptsFromRequest: Parsed options: limit=%d, page=%d, week=%d, month=%d, year=%d, from=%d, to=%d, artist_id=%d, album_id=%d, track_id=%d, period=%s",
		limit, page, tf.Week, tf.Month, tf.Year, tf.FromUnix, tf.ToUnix, artistId, albumId, trackId, period)

	return db.GetItemsOpts{
		Limit:     limit,
		Page:      page,
		Timeframe: tf,
		ArtistID:  artistId,
		AlbumID:   albumId,
		TrackID:   trackId,
	}
}

func TimeframeFromRequest(r *http.Request) db.Timeframe {
	q := r.URL.Query()

	parseInt := func(key string) int {
		v := q.Get(key)
		if v == "" {
			return 0
		}
		i, _ := strconv.Atoi(v)
		return i
	}

	parseInt64 := func(key string) int64 {
		v := q.Get(key)
		if v == "" {
			return 0
		}
		i, _ := strconv.ParseInt(v, 10, 64)
		return i
	}

	return db.Timeframe{
		Period:   db.Period(q.Get("period")),
		Year:     parseInt("year"),
		Month:    parseInt("month"),
		Week:     parseInt("week"),
		FromUnix: parseInt64("from"),
		ToUnix:   parseInt64("to"),
		Timezone: parseTZ(r),
	}
}

func parseTZ(r *http.Request) *time.Location {

	// this map is obviously AI.
	// i manually referenced as many links as I could and couldn't find any
	// incorrect entries here so hopefully it is all correct.
	overrides := map[string]string{
		// --- North America ---
		"America/Indianapolis":  "America/Indiana/Indianapolis",
		"America/Knoxville":     "America/Indiana/Knoxville",
		"America/Louisville":    "America/Kentucky/Louisville",
		"America/Montreal":      "America/Toronto",
		"America/Shiprock":      "America/Denver",
		"America/Fort_Wayne":    "America/Indiana/Indianapolis",
		"America/Virgin":        "America/Port_of_Spain",
		"America/Santa_Isabel":  "America/Tijuana",
		"America/Ensenada":      "America/Tijuana",
		"America/Rosario":       "America/Argentina/Cordoba",
		"America/Jujuy":         "America/Argentina/Jujuy",
		"America/Mendoza":       "America/Argentina/Mendoza",
		"America/Catamarca":     "America/Argentina/Catamarca",
		"America/Cordoba":       "America/Argentina/Cordoba",
		"America/Buenos_Aires":  "America/Argentina/Buenos_Aires",
		"America/Coral_Harbour": "America/Atikokan",
		"America/Atka":          "America/Adak",
		"US/Alaska":             "America/Anchorage",
		"US/Aleutian":           "America/Adak",
		"US/Arizona":            "America/Phoenix",
		"US/Central":            "America/Chicago",
		"US/Eastern":            "America/New_York",
		"US/East-Indiana":       "America/Indiana/Indianapolis",
		"US/Hawaii":             "Pacific/Honolulu",
		"US/Indiana-Starke":     "America/Indiana/Knoxville",
		"US/Michigan":           "America/Detroit",
		"US/Mountain":           "America/Denver",
		"US/Pacific":            "America/Los_Angeles",
		"US/Samoa":              "Pacific/Pago_Pago",
		"Canada/Atlantic":       "America/Halifax",
		"Canada/Central":        "America/Winnipeg",
		"Canada/Eastern":        "America/Toronto",
		"Canada/Mountain":       "America/Edmonton",
		"Canada/Newfoundland":   "America/St_Johns",
		"Canada/Pacific":        "America/Vancouver",

		// --- Asia ---
		"Asia/Calcutta":      "Asia/Kolkata",
		"Asia/Saigon":        "Asia/Ho_Chi_Minh",
		"Asia/Katmandu":      "Asia/Kathmandu",
		"Asia/Rangoon":       "Asia/Yangon",
		"Asia/Ulan_Bator":    "Asia/Ulaanbaatar",
		"Asia/Macao":         "Asia/Macau",
		"Asia/Tel_Aviv":      "Asia/Jerusalem",
		"Asia/Ashkhabad":     "Asia/Ashgabat",
		"Asia/Chungking":     "Asia/Chongqing",
		"Asia/Dacca":         "Asia/Dhaka",
		"Asia/Istanbul":      "Europe/Istanbul",
		"Asia/Kashgar":       "Asia/Urumqi",
		"Asia/Thimbu":        "Asia/Thimphu",
		"Asia/Ujung_Pandang": "Asia/Makassar",
		"ROC":                "Asia/Taipei",
		"Iran":               "Asia/Tehran",
		"Israel":             "Asia/Jerusalem",
		"Japan":              "Asia/Tokyo",
		"Singapore":          "Asia/Singapore",
		"Hongkong":           "Asia/Hong_Kong",

		// --- Europe ---
		"Europe/Kiev":     "Europe/Kyiv",
		"Europe/Belfast":  "Europe/London",
		"Europe/Tiraspol": "Europe/Chisinau",
		"Europe/Nicosia":  "Asia/Nicosia",
		"Europe/Moscow":   "Europe/Moscow",
		"W-SU":            "Europe/Moscow",
		"GB":              "Europe/London",
		"GB-Eire":         "Europe/London",
		"Eire":            "Europe/Dublin",
		"Poland":          "Europe/Warsaw",
		"Portugal":        "Europe/Lisbon",
		"Turkey":          "Europe/Istanbul",

		// --- Australia / Pacific ---
		"Australia/ACT":        "Australia/Sydney",
		"Australia/Canberra":   "Australia/Sydney",
		"Australia/LHI":        "Australia/Lord_Howe",
		"Australia/North":      "Australia/Darwin",
		"Australia/NSW":        "Australia/Sydney",
		"Australia/Queensland": "Australia/Brisbane",
		"Australia/South":      "Australia/Adelaide",
		"Australia/Tasmania":   "Australia/Hobart",
		"Australia/Victoria":   "Australia/Melbourne",
		"Australia/West":       "Australia/Perth",
		"Australia/Yancowinna": "Australia/Broken_Hill",
		"Pacific/Samoa":        "Pacific/Pago_Pago",
		"Pacific/Yap":          "Pacific/Chuuk",
		"Pacific/Truk":         "Pacific/Chuuk",
		"Pacific/Ponape":       "Pacific/Pohnpei",
		"NZ":                   "Pacific/Auckland",
		"NZ-CHAT":              "Pacific/Chatham",

		// --- Africa ---
		"Africa/Asmera":   "Africa/Asmara",
		"Africa/Timbuktu": "Africa/Bamako",
		"Egypt":           "Africa/Cairo",
		"Libya":           "Africa/Tripoli",

		// --- Atlantic ---
		"Atlantic/Faeroe":    "Atlantic/Faroe",
		"Atlantic/Jan_Mayen": "Europe/Oslo",
		"Iceland":            "Atlantic/Reykjavik",

		// --- Etc / Misc ---
		"UTC":       "UTC",
		"Etc/UTC":   "UTC",
		"Etc/GMT":   "UTC",
		"GMT":       "UTC",
		"Zulu":      "UTC",
		"Universal": "UTC",
	}

	if cfg.ForceTZ() != nil {
		return cfg.ForceTZ()
	}

	if tz := r.URL.Query().Get("tz"); tz != "" {
		if fixedTz, exists := overrides[tz]; exists {
			tz = fixedTz
		}
		if loc, err := time.LoadLocation(tz); err == nil {
			return loc
		}
	}

	if c, err := r.Cookie("tz"); err == nil {
		var tz string
		if fixedTz, exists := overrides[c.Value]; exists {
			tz = fixedTz
		} else {
			tz = c.Value
		}
		if loc, err := time.LoadLocation(tz); err == nil {
			return loc
		}
	}

	return time.Now().Location()
}
