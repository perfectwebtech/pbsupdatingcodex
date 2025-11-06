package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"
)

// ============================================================================
// AI-Powered Recommendation System
// ============================================================================
// Features:
// - Collaborative filtering (user-based and item-based)
// - Content-based filtering (genre, category, tags)
// - Hybrid recommendations combining multiple algorithms
// - Trending content detection
// - Personalized content discovery
// - Real-time recommendation updates
// ============================================================================

type AIRecommendationsService struct {
	db                  *sql.DB
	minRecommendations  int
	maxRecommendations  int
	trendingWindowHours int
	cacheDuration       time.Duration
}

// Recommendation represents a recommended content item
type Recommendation struct {
	ContentID   int64   `json:"content_id"`
	ContentType string  `json:"content_type"` // stream, movie, series
	Title       string  `json:"title"`
	Poster      string  `json:"poster"`
	Score       float64 `json:"score"`       // Recommendation confidence (0-100)
	Reason      string  `json:"reason"`      // Why recommended
	Tags        []string `json:"tags"`
	Category    string  `json:"category"`
	Rating      float64 `json:"rating"`
}

// TrendingContent represents trending content
type TrendingContent struct {
	ContentID     int64   `json:"content_id"`
	ContentType   string  `json:"content_type"`
	Title         string  `json:"title"`
	Poster        string  `json:"poster"`
	TrendScore    float64 `json:"trend_score"`
	ViewCount     int     `json:"view_count"`
	GrowthRate    float64 `json:"growth_rate"` // Percentage growth
	Category      string  `json:"category"`
}

// UserPreferences represents learned user preferences
type UserPreferences struct {
	UserID            int64              `json:"user_id"`
	FavoriteGenres    []string           `json:"favorite_genres"`
	FavoriteCategories []int64           `json:"favorite_categories"`
	PreferredQuality  string             `json:"preferred_quality"`
	PreferredLanguage string             `json:"preferred_language"`
	WatchTimeProfile  map[string]float64 `json:"watch_time_profile"` // hour -> percentage
	ContentTypeRatio  map[string]float64 `json:"content_type_ratio"` // live/vod/series
}

// NewAIRecommendationsService creates a new AI recommendations service
func NewAIRecommendationsService(db *sql.DB) *AIRecommendationsService {
	return &AIRecommendationsService{
		db:                  db,
		minRecommendations:  5,
		maxRecommendations:  50,
		trendingWindowHours: 24,
		cacheDuration:       15 * time.Minute,
	}
}

// ============================================================================
// Main Recommendation Methods
// ============================================================================

// GetPersonalizedRecommendations returns personalized content recommendations
func (s *AIRecommendationsService) GetPersonalizedRecommendations(userID int64, limit int) ([]Recommendation, error) {
	if limit == 0 {
		limit = 20
	}

	// 1. Get user preferences
	prefs, err := s.GetUserPreferences(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}

	// 2. Generate recommendations from multiple sources
	collaborativeRecs := s.getCollaborativeRecommendations(userID, limit)
	contentBasedRecs := s.getContentBasedRecommendations(userID, prefs, limit)
	trendingRecs := s.getTrendingRecommendations(userID, limit/4)

	// 3. Combine and deduplicate recommendations
	allRecs := s.mergeRecommendations(collaborativeRecs, contentBasedRecs, trendingRecs)

	// 4. Filter out already watched content
	filteredRecs := s.filterWatchedContent(userID, allRecs)

	// 5. Sort by score and limit
	sort.Slice(filteredRecs, func(i, j int) bool {
		return filteredRecs[i].Score > filteredRecs[j].Score
	})

	if len(filteredRecs) > limit {
		filteredRecs = filteredRecs[:limit]
	}

	return filteredRecs, nil
}

// GetTrendingContent returns currently trending content
func (s *AIRecommendationsService) GetTrendingContent(limit int, category string) ([]TrendingContent, error) {
	if limit == 0 {
		limit = 20
	}

	// Calculate trending score based on view velocity and total views
	query := `
		SELECT
			v.content_id,
			v.content_type,
			COALESCE(s.name, m.title, ser.title) as title,
			COALESCE(s.logo, m.poster, ser.poster) as poster,
			COUNT(v.id) as view_count,
			v.category,
			AVG(COALESCE(r.rating, 0)) as avg_rating,
			(COUNT(v.id) * (1 + (COUNT(v.id) / (1 + TIMESTAMPDIFF(HOUR, MIN(v.created_at), NOW()))))) as trend_score
		FROM view_history v
		LEFT JOIN streams s ON v.content_type = 'stream' AND v.content_id = s.id
		LEFT JOIN vod_movies m ON v.content_type = 'movie' AND v.content_id = m.id
		LEFT JOIN vod_series ser ON v.content_type = 'series' AND v.content_id = ser.id
		LEFT JOIN ratings r ON r.content_type = v.content_type AND r.content_id = v.content_id
		WHERE v.created_at >= DATE_SUB(NOW(), INTERVAL ? HOUR)
		%s
		GROUP BY v.content_id, v.content_type
		HAVING view_count >= 3
		ORDER BY trend_score DESC
		LIMIT ?
	`

	var categoryFilter string
	var args []interface{}
	args = append(args, s.trendingWindowHours)

	if category != "" {
		categoryFilter = "AND v.category = ?"
		args = append(args, category)
	}
	args = append(args, limit)

	query = fmt.Sprintf(query, categoryFilter)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query trending content: %w", err)
	}
	defer rows.Close()

	var trending []TrendingContent
	for rows.Next() {
		var item TrendingContent
		var avgRating sql.NullFloat64

		err := rows.Scan(
			&item.ContentID, &item.ContentType, &item.Title, &item.Poster,
			&item.ViewCount, &item.Category, &avgRating, &item.TrendScore,
		)
		if err != nil {
			continue
		}

		// Calculate growth rate (views in last 6 hours vs previous 6 hours)
		item.GrowthRate = s.calculateGrowthRate(item.ContentID, item.ContentType)

		trending = append(trending, item)
	}

	return trending, nil
}

// GetUserPreferences analyzes user behavior and returns preferences
func (s *AIRecommendationsService) GetUserPreferences(userID int64) (*UserPreferences, error) {
	prefs := &UserPreferences{
		UserID:           userID,
		WatchTimeProfile: make(map[string]float64),
		ContentTypeRatio: make(map[string]float64),
	}

	// 1. Analyze favorite genres from watch history
	prefs.FavoriteGenres = s.analyzeFavoriteGenres(userID)

	// 2. Analyze favorite categories
	prefs.FavoriteCategories = s.analyzeFavoriteCategories(userID)

	// 3. Analyze preferred quality
	prefs.PreferredQuality = s.analyzePreferredQuality(userID)

	// 4. Analyze watch time patterns
	prefs.WatchTimeProfile = s.analyzeWatchTimeProfile(userID)

	// 5. Analyze content type preferences
	prefs.ContentTypeRatio = s.analyzeContentTypeRatio(userID)

	return prefs, nil
}

// GetSimilarContent returns content similar to a given item
func (s *AIRecommendationsService) GetSimilarContent(contentType string, contentID int64, limit int) ([]Recommendation, error) {
	if limit == 0 {
		limit = 10
	}

	var recommendations []Recommendation

	switch contentType {
	case "stream":
		recommendations = s.getSimilarStreams(contentID, limit)
	case "movie":
		recommendations = s.getSimilarMovies(contentID, limit)
	case "series":
		recommendations = s.getSimilarSeries(contentID, limit)
	default:
		return nil, fmt.Errorf("invalid content type: %s", contentType)
	}

	return recommendations, nil
}

// ============================================================================
// Collaborative Filtering
// ============================================================================

func (s *AIRecommendationsService) getCollaborativeRecommendations(userID int64, limit int) []Recommendation {
	// Find similar users based on watch history overlap
	similarUsers := s.findSimilarUsers(userID, 20)

	// Get content watched by similar users but not by current user
	query := `
		SELECT DISTINCT
			v.content_id,
			v.content_type,
			COALESCE(s.name, m.title, ser.title) as title,
			COALESCE(s.logo, m.poster, ser.poster) as poster,
			COUNT(v.user_id) as watch_count,
			AVG(COALESCE(r.rating, 3.5)) as avg_rating
		FROM view_history v
		LEFT JOIN streams s ON v.content_type = 'stream' AND v.content_id = s.id
		LEFT JOIN vod_movies m ON v.content_type = 'movie' AND v.content_id = m.id
		LEFT JOIN vod_series ser ON v.content_type = 'series' AND v.content_id = ser.id
		LEFT JOIN ratings r ON r.content_type = v.content_type AND r.content_id = v.content_id
		WHERE v.user_id IN (?)
			AND NOT EXISTS (
				SELECT 1 FROM view_history vh
				WHERE vh.user_id = ? AND vh.content_id = v.content_id AND vh.content_type = v.content_type
			)
		GROUP BY v.content_id, v.content_type
		ORDER BY watch_count DESC, avg_rating DESC
		LIMIT ?
	`

	if len(similarUsers) == 0 {
		return []Recommendation{}
	}

	userIDsStr := fmt.Sprintf("%d", similarUsers[0])
	for i := 1; i < len(similarUsers); i++ {
		userIDsStr += fmt.Sprintf(",%d", similarUsers[i])
	}

	rows, err := s.db.Query(query, userIDsStr, userID, limit)
	if err != nil {
		return []Recommendation{}
	}
	defer rows.Close()

	var recommendations []Recommendation
	for rows.Next() {
		var rec Recommendation
		var watchCount int
		rows.Scan(&rec.ContentID, &rec.ContentType, &rec.Title, &rec.Poster, &watchCount, &rec.Rating)

		// Score based on how many similar users watched it
		rec.Score = (float64(watchCount) / float64(len(similarUsers))) * 100
		rec.Reason = "Users like you watched this"

		recommendations = append(recommendations, rec)
	}

	return recommendations
}

func (s *AIRecommendationsService) findSimilarUsers(userID int64, limit int) []int64 {
	// Jaccard similarity based on watch history
	query := `
		SELECT
			v2.user_id,
			COUNT(DISTINCT CASE WHEN v1.content_id = v2.content_id AND v1.content_type = v2.content_type THEN v1.content_id END) as common_items,
			COUNT(DISTINCT v1.content_id) + COUNT(DISTINCT v2.content_id) - COUNT(DISTINCT CASE WHEN v1.content_id = v2.content_id THEN v1.content_id END) as total_items
		FROM view_history v1
		CROSS JOIN view_history v2
		WHERE v1.user_id = ? AND v2.user_id != ? AND v2.user_id IS NOT NULL
		GROUP BY v2.user_id
		HAVING common_items > 0
		ORDER BY (common_items * 1.0 / total_items) DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, userID, userID, limit)
	if err != nil {
		return []int64{}
	}
	defer rows.Close()

	var similarUsers []int64
	for rows.Next() {
		var uid int64
		var commonItems, totalItems int
		rows.Scan(&uid, &commonItems, &totalItems)
		similarUsers = append(similarUsers, uid)
	}

	return similarUsers
}

// ============================================================================
// Content-Based Filtering
// ============================================================================

func (s *AIRecommendationsService) getContentBasedRecommendations(userID int64, prefs *UserPreferences, limit int) []Recommendation {
	var recommendations []Recommendation

	// Build query based on user preferences
	genresFilter := ""
	categoriesFilter := ""

	if len(prefs.FavoriteGenres) > 0 {
		genresFilter = fmt.Sprintf("AND (genres LIKE '%%%s%%'", prefs.FavoriteGenres[0])
		for i := 1; i < len(prefs.FavoriteGenres); i++ {
			genresFilter += fmt.Sprintf(" OR genres LIKE '%%%s%%'", prefs.FavoriteGenres[i])
		}
		genresFilter += ")"
	}

	if len(prefs.FavoriteCategories) > 0 {
		categoriesFilter = fmt.Sprintf("AND category_id IN (%d", prefs.FavoriteCategories[0])
		for i := 1; i < len(prefs.FavoriteCategories); i++ {
			categoriesFilter += fmt.Sprintf(",%d", prefs.FavoriteCategories[i])
		}
		categoriesFilter += ")"
	}

	// Get movies matching preferences
	movieQuery := fmt.Sprintf(`
		SELECT
			id as content_id,
			'movie' as content_type,
			title,
			poster,
			rating,
			genres
		FROM vod_movies
		WHERE is_active = 1
			%s
			%s
			AND NOT EXISTS (
				SELECT 1 FROM view_history WHERE user_id = ? AND content_id = id AND content_type = 'movie'
			)
		ORDER BY rating DESC, created_at DESC
		LIMIT ?
	`, genresFilter, categoriesFilter)

	rows, err := s.db.Query(movieQuery, userID, limit/2)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var rec Recommendation
			var genresJSON string
			rows.Scan(&rec.ContentID, &rec.ContentType, &rec.Title, &rec.Poster, &rec.Rating, &genresJSON)

			json.Unmarshal([]byte(genresJSON), &rec.Tags)
			rec.Score = s.calculateContentScore(rec, prefs)
			rec.Reason = "Matches your interests"

			recommendations = append(recommendations, rec)
		}
	}

	return recommendations
}

func (s *AIRecommendationsService) calculateContentScore(rec Recommendation, prefs *UserPreferences) float64 {
	score := 50.0 // Base score

	// Boost score based on genre match
	genreMatches := 0
	for _, tag := range rec.Tags {
		for _, favGenre := range prefs.FavoriteGenres {
			if tag == favGenre {
				genreMatches++
			}
		}
	}
	score += float64(genreMatches) * 10

	// Boost based on rating
	score += rec.Rating * 5

	// Cap at 100
	if score > 100 {
		score = 100
	}

	return score
}

// ============================================================================
// Trending Recommendations
// ============================================================================

func (s *AIRecommendationsService) getTrendingRecommendations(userID int64, limit int) []Recommendation {
	trending, err := s.GetTrendingContent(limit, "")
	if err != nil {
		return []Recommendation{}
	}

	var recommendations []Recommendation
	for _, item := range trending {
		rec := Recommendation{
			ContentID:   item.ContentID,
			ContentType: item.ContentType,
			Title:       item.Title,
			Poster:      item.Poster,
			Score:       math.Min(item.TrendScore/10, 100), // Normalize to 0-100
			Reason:      "Trending now",
			Category:    item.Category,
		}
		recommendations = append(recommendations, rec)
	}

	return recommendations
}

// ============================================================================
// Similar Content (Item-to-Item)
// ============================================================================

func (s *AIRecommendationsService) getSimilarStreams(streamID int64, limit int) []Recommendation {
	query := `
		SELECT
			s2.id,
			s2.name,
			s2.logo,
			s2.category_id,
			COUNT(DISTINCT v.user_id) as common_viewers
		FROM streams s1
		JOIN view_history v1 ON v1.content_id = s1.id AND v1.content_type = 'stream'
		JOIN view_history v2 ON v2.user_id = v1.user_id AND v2.content_type = 'stream'
		JOIN streams s2 ON s2.id = v2.content_id
		WHERE s1.id = ? AND s2.id != ? AND s2.is_active = 1
		GROUP BY s2.id
		ORDER BY common_viewers DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, streamID, streamID, limit)
	if err != nil {
		return []Recommendation{}
	}
	defer rows.Close()

	var recommendations []Recommendation
	for rows.Next() {
		var rec Recommendation
		var categoryID int64
		var commonViewers int
		rows.Scan(&rec.ContentID, &rec.Title, &rec.Poster, &categoryID, &commonViewers)

		rec.ContentType = "stream"
		rec.Score = float64(commonViewers) * 2 // Simple scoring
		rec.Reason = "Similar to what you watched"

		recommendations = append(recommendations, rec)
	}

	return recommendations
}

func (s *AIRecommendationsService) getSimilarMovies(movieID int64, limit int) []Recommendation {
	// Get movie genres
	var genresJSON string
	s.db.QueryRow("SELECT genres FROM vod_movies WHERE id = ?", movieID).Scan(&genresJSON)

	var genres []string
	json.Unmarshal([]byte(genresJSON), &genres)

	if len(genres) == 0 {
		return []Recommendation{}
	}

	// Find movies with similar genres
	genreFilter := fmt.Sprintf("(genres LIKE '%%%s%%'", genres[0])
	for i := 1; i < len(genres); i++ {
		genreFilter += fmt.Sprintf(" OR genres LIKE '%%%s%%'", genres[i])
	}
	genreFilter += ")"

	query := fmt.Sprintf(`
		SELECT id, title, poster, rating, genres
		FROM vod_movies
		WHERE %s AND id != ? AND is_active = 1
		ORDER BY rating DESC
		LIMIT ?
	`, genreFilter)

	rows, err := s.db.Query(query, movieID, limit)
	if err != nil {
		return []Recommendation{}
	}
	defer rows.Close()

	var recommendations []Recommendation
	for rows.Next() {
		var rec Recommendation
		rows.Scan(&rec.ContentID, &rec.Title, &rec.Poster, &rec.Rating, &genresJSON)

		json.Unmarshal([]byte(genresJSON), &rec.Tags)
		rec.ContentType = "movie"
		rec.Score = rec.Rating * 10
		rec.Reason = "Similar genre"

		recommendations = append(recommendations, rec)
	}

	return recommendations
}

func (s *AIRecommendationsService) getSimilarSeries(seriesID int64, limit int) []Recommendation {
	// Similar implementation to movies
	return s.getSimilarMovies(seriesID, limit) // Reuse logic for now
}

// ============================================================================
// User Preference Analysis
// ============================================================================

func (s *AIRecommendationsService) analyzeFavoriteGenres(userID int64) []string {
	query := `
		SELECT genres, COUNT(*) as count
		FROM (
			SELECT m.genres FROM view_history v
			JOIN vod_movies m ON v.content_id = m.id AND v.content_type = 'movie'
			WHERE v.user_id = ?
			UNION ALL
			SELECT ser.genres FROM view_history v
			JOIN vod_series ser ON v.content_id = ser.id AND v.content_type = 'series'
			WHERE v.user_id = ?
		) as all_genres
		GROUP BY genres
		ORDER BY count DESC
		LIMIT 5
	`

	rows, err := s.db.Query(query, userID, userID)
	if err != nil {
		return []string{}
	}
	defer rows.Close()

	genreMap := make(map[string]int)
	for rows.Next() {
		var genresJSON string
		var count int
		rows.Scan(&genresJSON, &count)

		var genres []string
		json.Unmarshal([]byte(genresJSON), &genres)

		for _, genre := range genres {
			genreMap[genre] += count
		}
	}

	// Sort by count
	type genreCount struct {
		genre string
		count int
	}
	var genreCounts []genreCount
	for genre, count := range genreMap {
		genreCounts = append(genreCounts, genreCount{genre, count})
	}
	sort.Slice(genreCounts, func(i, j int) bool {
		return genreCounts[i].count > genreCounts[j].count
	})

	var favoriteGenres []string
	for i := 0; i < len(genreCounts) && i < 5; i++ {
		favoriteGenres = append(favoriteGenres, genreCounts[i].genre)
	}

	return favoriteGenres
}

func (s *AIRecommendationsService) analyzeFavoriteCategories(userID int64) []int64 {
	query := `
		SELECT category_id, COUNT(*) as count
		FROM view_history
		WHERE user_id = ? AND category_id IS NOT NULL
		GROUP BY category_id
		ORDER BY count DESC
		LIMIT 3
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return []int64{}
	}
	defer rows.Close()

	var categories []int64
	for rows.Next() {
		var categoryID int64
		var count int
		rows.Scan(&categoryID, &count)
		categories = append(categories, categoryID)
	}

	return categories
}

func (s *AIRecommendationsService) analyzePreferredQuality(userID int64) string {
	query := `
		SELECT quality, COUNT(*) as count
		FROM view_history
		WHERE user_id = ? AND quality IS NOT NULL
		GROUP BY quality
		ORDER BY count DESC
		LIMIT 1
	`

	var quality string
	var count int
	err := s.db.QueryRow(query, userID).Scan(&quality, &count)
	if err != nil {
		return "hd" // Default
	}

	return quality
}

func (s *AIRecommendationsService) analyzeWatchTimeProfile(userID int64) map[string]float64 {
	profile := make(map[string]float64)

	query := `
		SELECT HOUR(created_at) as hour, COUNT(*) as count
		FROM view_history
		WHERE user_id = ?
		GROUP BY hour
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return profile
	}
	defer rows.Close()

	totalViews := 0
	hourCounts := make(map[int]int)

	for rows.Next() {
		var hour, count int
		rows.Scan(&hour, &count)
		hourCounts[hour] = count
		totalViews += count
	}

	// Convert to percentages
	for hour, count := range hourCounts {
		profile[fmt.Sprintf("%02d:00", hour)] = (float64(count) / float64(totalViews)) * 100
	}

	return profile
}

func (s *AIRecommendationsService) analyzeContentTypeRatio(userID int64) map[string]float64 {
	ratio := make(map[string]float64)

	query := `
		SELECT content_type, COUNT(*) as count
		FROM view_history
		WHERE user_id = ?
		GROUP BY content_type
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return ratio
	}
	defer rows.Close()

	totalViews := 0
	typeCounts := make(map[string]int)

	for rows.Next() {
		var contentType string
		var count int
		rows.Scan(&contentType, &count)
		typeCounts[contentType] = count
		totalViews += count
	}

	// Convert to percentages
	for contentType, count := range typeCounts {
		ratio[contentType] = (float64(count) / float64(totalViews)) * 100
	}

	return ratio
}

// ============================================================================
// Helper Methods
// ============================================================================

func (s *AIRecommendationsService) calculateGrowthRate(contentID int64, contentType string) float64 {
	query := `
		SELECT
			(SELECT COUNT(*) FROM view_history WHERE content_id = ? AND content_type = ? AND created_at >= DATE_SUB(NOW(), INTERVAL 6 HOUR)) as recent,
			(SELECT COUNT(*) FROM view_history WHERE content_id = ? AND content_type = ? AND created_at BETWEEN DATE_SUB(NOW(), INTERVAL 12 HOUR) AND DATE_SUB(NOW(), INTERVAL 6 HOUR)) as previous
	`

	var recent, previous int
	err := s.db.QueryRow(query, contentID, contentType, contentID, contentType).Scan(&recent, &previous)
	if err != nil || previous == 0 {
		return 0
	}

	return ((float64(recent) - float64(previous)) / float64(previous)) * 100
}

func (s *AIRecommendationsService) mergeRecommendations(lists ...[]Recommendation) []Recommendation {
	seen := make(map[string]bool)
	var merged []Recommendation

	for _, list := range lists {
		for _, rec := range list {
			key := fmt.Sprintf("%s:%d", rec.ContentType, rec.ContentID)
			if !seen[key] {
				seen[key] = true
				merged = append(merged, rec)
			}
		}
	}

	return merged
}

func (s *AIRecommendationsService) filterWatchedContent(userID int64, recommendations []Recommendation) []Recommendation {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM view_history
			WHERE user_id = ? AND content_id = ? AND content_type = ?
		)
	`

	var filtered []Recommendation
	for _, rec := range recommendations {
		var exists bool
		s.db.QueryRow(query, userID, rec.ContentID, rec.ContentType).Scan(&exists)
		if !exists {
			filtered = append(filtered, rec)
		}
	}

	return filtered
}
