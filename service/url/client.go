package url

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/readfog/url/cache"
	"github.com/readfog/url/common"
	"github.com/readfog/url/model"
	"github.com/readfog/url/orm"
	"github.com/readfog/url/request"
	"github.com/readfog/url/util"
	"gorm.io/gorm"
)

// CreateURLShortCodeFromRequest creates a new short code for url given in http.Request
// It uses expires_on date and keywords from http.Request if available.
// It returns created short code or error if any.
func CreateURLShortCodeFromRequest(req *http.Request) (string, error) {
	var input request.URLInput

	if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
		return "", err
	}

	return CreateURLShortCode(input)
}

// CreateURLShortCode creates a new short code using request.URLInput
// It returns created short code or error if any.
func CreateURLShortCode(input request.URLInput) (string, error) {
	if shortCode, err := ValidateURLInput(input); err != nil {
		return shortCode, err
	}

	shortCode := getUniqueShortCode()
	expiresOn, _ := input.GetExpiresOn()

	orm.Connection().Create(&model.URL{
		ShortCode: shortCode,
		OriginURL: input.URL,
		Keywords:  mapKeywords(input.Keywords),
		ExpiresOn: expiresOn,
	})

	return shortCode, nil
}

// LookupOriginURL looks up origin url from shortCode
// It returns origin url if exists and is active, http error code otherwise.
func LookupOriginURL(shortCode string) (model.URL, int, bool) {
	var urlModel model.URL
	log.Printf("LookupOriginURL shortCode=%v", shortCode)
	if cacheModel, status := cache.LookupURL(shortCode); status != 0 {
		log.Print("cache")
		return cacheModel, status, true
	}

	if status := orm.Connection().Where("short_code = ?", shortCode).First(&urlModel); status.RowsAffected == 0 {
		log.Print("from db")
		return urlModel, http.StatusNotFound, false
	}

	if !urlModel.IsActive() {
		if !urlModel.Deleted {
			go cache.DeactivateURL(urlModel)
		}

		return urlModel, http.StatusGone, false
	}

	return urlModel, http.StatusMovedPermanently, false
}

// IncrementHits increments hit counter for given shortCode just before serving redirection
func IncrementHits(urlModel model.URL) {
	orm.Connection().Model(&urlModel).
		Where("short_code = ?", urlModel.ShortCode).
		UpdateColumn("hits", gorm.Expr("hits + ?", 1))

	if urlModel.Hits+1 >= common.PopularHits {
		cache.SavePopularURL(urlModel, false)
	}
}

// allowDupeURL checks is app is configured to allow dupe url
func allowDupeURL() bool {
	return os.Getenv("APP_ALLOW_DUPE_URL") == "1"
}

// checkURLReach checks is app is configured to check if origin url host is reachable
func checkURLReach() bool {
	return os.Getenv("APP_CHECK_URL_REACH") == "1"
}

// ValidateURLInput validates given request.URLInput
// It returns existing short code for input url if exists and validation error if any.
func ValidateURLInput(input request.URLInput) (string, error) {
	if err := input.Validate(); err != nil {
		return "", err
	}

	if !allowDupeURL() {
		if shortCode := getShortCodeByOriginURL(input.URL); shortCode != "" {
			return shortCode, common.ErrURLAlreadyShort
		}
	}

	if input.Host == "" || !checkURLReach() {
		return "", nil
	}

	if _, err := net.LookupIP(input.Host); err != nil {
		return "", common.ErrInvalidURL
	}

	return "", nil
}

// getUniqueShortCode gets unique random string to use as short code
// It checks db to ensure it is really unique and returns short code string.
func getUniqueShortCode() string {
	shortCode := util.RandomString(common.ShortCodeLength)

	for {
		if !isExistingShortCode(shortCode) {
			return shortCode
		}

		shortCode = util.RandomString(common.ShortCodeLength)
	}
}

// isExistingShortCode checks if given short code exists in db
// It returns bool.
func isExistingShortCode(shortCode string) bool {
	var urlModel model.URL

	status := orm.Connection().Select("short_code").Where("short_code = ?", shortCode).First(&urlModel)

	return status.RowsAffected > 0
}

// getShortCodeByOriginURL gets short code for given origin url
// It returns short code string.
func getShortCodeByOriginURL(originURL string) string {
	var urlModel model.URL

	orm.Connection().Select("short_code").
		Where("origin_url = ? AND deleted = ?", originURL, false).
		First(&urlModel)

	return urlModel.ShortCode
}

// mapKeywords maps input keyword array to model arrays
// It returns array of model.Keyword.
func mapKeywords(words []string) []model.Keyword {
	var Keywords []model.Keyword

	for _, word := range words {
		var keyword model.Keyword
		orm.Connection().FirstOrCreate(&keyword, model.Keyword{Keyword: word})
		Keywords = append(Keywords, keyword)
	}

	return Keywords
}
