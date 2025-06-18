package utils

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gabehf/koito/internal/mbz"
	"github.com/gabehf/koito/internal/models"
	"github.com/google/uuid"
)

func IDFromString(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	return s
}

func ParseUUIDSlice(str []string) ([]uuid.UUID, error) {
	ret := make([]uuid.UUID, 0)
	for _, s := range str {
		parsed, err := uuid.Parse(s)
		if err != nil {
			continue
		}
		ret = append(ret, parsed)
	}
	return ret, nil
}

func FlattenArtistMbzIDs(artists []*models.Artist) []uuid.UUID {
	ids := make([]uuid.UUID, 0)
	for _, a := range artists {
		if a.MbzID == nil || *a.MbzID == uuid.Nil {
			continue
		}
		ids = append(ids, *a.MbzID)
	}
	return ids
}

func FlattenArtistNames(artists []*models.Artist) []string {
	names := make([]string, 0)
	for _, a := range artists {
		names = append(names, a.Aliases...)
	}
	return names
}

func FlattenSimpleArtistNames(artists []models.SimpleArtist) []string {
	names := make([]string, 0)
	for _, a := range artists {
		names = append(names, a.Name)
	}
	return names
}

func FlattenMbzArtistCreditNames(artists []mbz.MusicBrainzArtistCredit) []string {
	names := make([]string, len(artists))
	for i, a := range artists {
		names[i] = a.Name
	}
	return names
}

func FlattenArtistIDs(artists []*models.Artist) []int32 {
	ids := make([]int32, len(artists))
	for i, a := range artists {
		ids[i] = a.ID
	}
	return ids
}

// DateRange takes optional week, month, and year. If all are 0, it returns the zero time range.
// If only year is provided, it returns the full year.
// If both month and year are provided, it returns the start and end of that month.
// If week and year are provided, it returns the start and end of that week.
// If only week or month is provided without a year, it's considered invalid.
func DateRange(week, month, year int) (time.Time, time.Time, error) {
	if week == 0 && month == 0 && year == 0 {
		// No filter applied
		return time.Time{}, time.Time{}, nil
	}

	if month != 0 && (month < 1 || month > 12) {
		return time.Time{}, time.Time{}, errors.New("DateRange: invalid month")
	}

	if week != 0 && (week < 1 || week > 53) {
		return time.Time{}, time.Time{}, errors.New("DateRange: invalid week")
	}

	if year < 1 {
		return time.Time{}, time.Time{}, errors.New("DateRange: invalid year")
	}

	loc := time.Local

	if week != 0 {
		if month != 0 {
			return time.Time{}, time.Time{}, errors.New("DateRange: cannot specify both week and month")
		}
		// Specific week
		start := time.Date(year, 1, 1, 0, 0, 0, 0, loc)
		start = start.AddDate(0, 0, (week-1)*7)
		end := start.AddDate(0, 0, 7)
		return start, end, nil
	}

	if month == 0 {
		// Whole year
		start := time.Date(year, 1, 1, 0, 0, 0, 0, loc)
		end := start.AddDate(1, 0, 0)
		return start, end, nil
	}

	// Specific month
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, loc)
	end := start.AddDate(0, 1, 0)
	return start, end, nil
}

// CopyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success. Otherise, attempt to create a hard link
// between the two files. If that fail, copy the file contents from src to dst.
func CopyFile(src, dst string) (err error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("CopyFile: %w", err)
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("CopyFile: %w", err)
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return fmt.Errorf("CopyFile: %w", err)
		}
	}
	if err = os.Link(src, dst); err == nil {
		return fmt.Errorf("CopyFile: %w", err)
	}
	err = copyFileContents(src, dst)
	if err != nil {
		return fmt.Errorf("CopyFile: %w", err)
	}
	return nil
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("copyFileContents: %w", err)
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("copyFileContents: %w", err)
	}
	defer out.Close()
	if _, err = io.Copy(out, in); err != nil {
		return fmt.Errorf("copyFileContents: %w", err)
	}
	err = out.Sync()
	if err != nil {
		return fmt.Errorf("copyFileContents: %w", err)
	}
	return nil
}

// Returns the same slice, but with all strings that are equal (with strings.EqualFold)
// included only once
func UniqueIgnoringCase(s []string) []string {
	unique := []string{}

	for _, str := range s {
		isDuplicate := false
		for _, u := range unique {
			if strings.EqualFold(str, u) {
				isDuplicate = true
				break
			}
		}
		if !isDuplicate {
			unique = append(unique, str)
		}
	}

	return unique
}

// Removes duplicates in a string set
func Unique(xs *[]string) {
	found := make(map[string]bool)
	j := 0
	for i, x := range *xs {
		if !found[x] {
			found[x] = true
			(*xs)[j] = (*xs)[i]
			j++
		}
	}
	*xs = (*xs)[:j]
}

// Returns the same slice, but with all entries that contain non ASCII characters removed
func RemoveNonAscii(s []string) []string {
	filtered := []string{}
	for _, str := range s {
		isAscii := true
		for _, r := range str {
			if r > 127 {
				isAscii = false
				break
			}
		}
		if isAscii {
			filtered = append(filtered, str)
		}
	}
	return filtered
}

// Returns only items that are in one slice but not the other
func RemoveInBoth(s, c []string) []string {
	result := []string{}
	set := make(map[string]struct{})

	for _, str := range c {
		set[str] = struct{}{}
	}

	for _, str := range s {
		if _, exists := set[str]; !exists {
			result = append(result, str)
		}
	}

	return result
}

// MoveFirstMatchToFront moves the first string containing the substring to the front of the slice.
func MoveFirstMatchToFront(slice []string, substring string) []string {
	for i, s := range slice {
		if strings.Contains(s, substring) {
			if i == 0 {
				return slice // already at the front
			}
			// Move the matching element to the front
			return append([]string{slice[i]}, append(slice[:i], slice[i+1:]...)...)
		}
	}
	// No match found, return unchanged
	return slice
}

// Taken with little modification from
// https://gist.github.com/dopey/c69559607800d2f2f90b1b1ed4e550fb?permalink_comment_id=3527095#gistcomment-3527095
func GenerateRandomString(length int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	ret := make([]byte, length)
	for i := range length {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", fmt.Errorf("GenerateRandomString: %w", err)
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}

// Essentially the same as utils.WriteError(w, `{"error": "message"}`, code)
func WriteError(w http.ResponseWriter, message string, code int) {
	http.Error(w, fmt.Sprintf(`{"error":"%s"}`, message), code)
}

// Sets content type and status code, and encodes data to json
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Returns true if more than one string is not empty
func MoreThanOneString(s ...string) bool {
	count := 0
	for _, str := range s {
		if str != "" {
			count++
		}
	}
	return count > 1
}

func ParseBool(s string) (value, ok bool) {
	if strings.ToLower(s) == "true" {
		value = true
		ok = true
		return
	} else if strings.ToLower(s) == "false" {
		value = false
		ok = true
		return
	} else {
		ok = false
		return
	}
}

func FlattenAliases(aliases []models.Alias) []string {
	ret := make([]string, len(aliases))
	for i := range aliases {
		ret[i] = aliases[i].Alias
	}
	return ret
}
