package graphql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/graphql-go/graphql"
)

// GraphQLService provides GraphQL API functionality
type GraphQLService struct {
	db     *sql.DB
	schema graphql.Schema
}

// NewGraphQLService creates a new GraphQL service
func NewGraphQLService(db *sql.DB) (*GraphQLService, error) {
	s := &GraphQLService{db: db}

	schema, err := s.buildSchema()
	if err != nil {
		return nil, err
	}

	s.schema = schema
	return s, nil
}

// =====================================================
// TYPE DEFINITIONS
// =====================================================

// User type
var userType = graphql.NewObject(graphql.ObjectConfig{
	Name: "User",
	Fields: graphql.Fields{
		"id":            &graphql.Field{Type: graphql.String},
		"username":      &graphql.Field{Type: graphql.String},
		"email":         &graphql.Field{Type: graphql.String},
		"avatar_url":    &graphql.Field{Type: graphql.String},
		"package_name":  &graphql.Field{Type: graphql.String},
		"is_active":     &graphql.Field{Type: graphql.Boolean},
		"created_at":    &graphql.Field{Type: graphql.DateTime},
		"subscription":  &graphql.Field{Type: subscriptionType},
		"watchlist":     &graphql.Field{Type: graphql.NewList(contentType)},
		"favorites":     &graphql.Field{Type: graphql.NewList(contentType)},
		"continue_watching": &graphql.Field{Type: graphql.NewList(contentType)},
	},
})

// Content interface (for streams, movies, series)
var contentInterface = graphql.NewInterface(graphql.InterfaceConfig{
	Name: "Content",
	Fields: graphql.Fields{
		"id":          &graphql.Field{Type: graphql.String},
		"title":       &graphql.Field{Type: graphql.String},
		"description": &graphql.Field{Type: graphql.String},
		"poster_url":  &graphql.Field{Type: graphql.String},
		"rating":      &graphql.Field{Type: graphql.Float},
		"duration":    &graphql.Field{Type: graphql.Int},
		"created_at":  &graphql.Field{Type: graphql.DateTime},
	},
	ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
		if obj, ok := p.Value.(map[string]interface{}); ok {
			if contentType, ok := obj["content_type"].(string); ok {
				switch contentType {
				case "stream":
					return streamType
				case "movie":
					return movieType
				case "series":
					return seriesType
				}
			}
		}
		return streamType
	},
})

// Stream type
var streamType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Stream",
	Interfaces: []*graphql.Interface{contentInterface},
	Fields: graphql.Fields{
		"id":          &graphql.Field{Type: graphql.String},
		"title":       &graphql.Field{Type: graphql.String},
		"description": &graphql.Field{Type: graphql.String},
		"poster_url":  &graphql.Field{Type: graphql.String},
		"rating":      &graphql.Field{Type: graphql.Float},
		"duration":    &graphql.Field{Type: graphql.Int},
		"created_at":  &graphql.Field{Type: graphql.DateTime},
		"stream_url":  &graphql.Field{Type: graphql.String},
		"category":    &graphql.Field{Type: categoryType},
		"is_live":     &graphql.Field{Type: graphql.Boolean},
		"viewers":     &graphql.Field{Type: graphql.Int},
		"epg":         &graphql.Field{Type: graphql.NewList(epgType)},
	},
})

// Movie type
var movieType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Movie",
	Interfaces: []*graphql.Interface{contentInterface},
	Fields: graphql.Fields{
		"id":          &graphql.Field{Type: graphql.String},
		"title":       &graphql.Field{Type: graphql.String},
		"description": &graphql.Field{Type: graphql.String},
		"poster_url":  &graphql.Field{Type: graphql.String},
		"rating":      &graphql.Field{Type: graphql.Float},
		"duration":    &graphql.Field{Type: graphql.Int},
		"created_at":  &graphql.Field{Type: graphql.DateTime},
		"genres":      &graphql.Field{Type: graphql.NewList(graphql.String)},
		"year":        &graphql.Field{Type: graphql.Int},
		"director":    &graphql.Field{Type: graphql.String},
		"cast":        &graphql.Field{Type: graphql.NewList(graphql.String)},
		"stream_url":  &graphql.Field{Type: graphql.String},
	},
})

// Series type
var seriesType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Series",
	Interfaces: []*graphql.Interface{contentInterface},
	Fields: graphql.Fields{
		"id":          &graphql.Field{Type: graphql.String},
		"title":       &graphql.Field{Type: graphql.String},
		"description": &graphql.Field{Type: graphql.String},
		"poster_url":  &graphql.Field{Type: graphql.String},
		"rating":      &graphql.Field{Type: graphql.Float},
		"duration":    &graphql.Field{Type: graphql.Int},
		"created_at":  &graphql.Field{Type: graphql.DateTime},
		"genres":      &graphql.Field{Type: graphql.NewList(graphql.String)},
		"seasons":     &graphql.Field{Type: graphql.NewList(seasonType)},
		"total_seasons": &graphql.Field{Type: graphql.Int},
	},
})

// Season type
var seasonType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Season",
	Fields: graphql.Fields{
		"season_number": &graphql.Field{Type: graphql.Int},
		"title":         &graphql.Field{Type: graphql.String},
		"episodes":      &graphql.Field{Type: graphql.NewList(episodeType)},
	},
})

// Episode type
var episodeType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Episode",
	Fields: graphql.Fields{
		"id":             &graphql.Field{Type: graphql.String},
		"episode_number": &graphql.Field{Type: graphql.Int},
		"title":          &graphql.Field{Type: graphql.String},
		"description":    &graphql.Field{Type: graphql.String},
		"duration":       &graphql.Field{Type: graphql.Int},
		"stream_url":     &graphql.Field{Type: graphql.String},
		"thumbnail_url":  &graphql.Field{Type: graphql.String},
	},
})

// Category type
var categoryType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Category",
	Fields: graphql.Fields{
		"id":          &graphql.Field{Type: graphql.String},
		"name":        &graphql.Field{Type: graphql.String},
		"description": &graphql.Field{Type: graphql.String},
		"icon":        &graphql.Field{Type: graphql.String},
		"order":       &graphql.Field{Type: graphql.Int},
	},
})

// Subscription type
var subscriptionType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Subscription",
	Fields: graphql.Fields{
		"package_name":  &graphql.Field{Type: graphql.String},
		"price":         &graphql.Field{Type: graphql.Float},
		"expires_at":    &graphql.Field{Type: graphql.DateTime},
		"is_trial":      &graphql.Field{Type: graphql.Boolean},
		"auto_renew":    &graphql.Field{Type: graphql.Boolean},
		"devices_limit": &graphql.Field{Type: graphql.Int},
	},
})

// EPG type
var epgType = graphql.NewObject(graphql.ObjectConfig{
	Name: "EPGEntry",
	Fields: graphql.Fields{
		"id":          &graphql.Field{Type: graphql.String},
		"title":       &graphql.Field{Type: graphql.String},
		"description": &graphql.Field{Type: graphql.String},
		"start_time":  &graphql.Field{Type: graphql.DateTime},
		"end_time":    &graphql.Field{Type: graphql.DateTime},
		"category":    &graphql.Field{Type: graphql.String},
	},
})

// Recommendation type
var recommendationType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Recommendation",
	Fields: graphql.Fields{
		"content":      &graphql.Field{Type: contentInterface},
		"reason":       &graphql.Field{Type: graphql.String},
		"score":        &graphql.Field{Type: graphql.Float},
		"personalized": &graphql.Field{Type: graphql.Boolean},
	},
})

// ContentType union
var contentType = graphql.NewUnion(graphql.UnionConfig{
	Name:  "ContentUnion",
	Types: []*graphql.Object{streamType, movieType, seriesType},
	ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
		if obj, ok := p.Value.(map[string]interface{}); ok {
			if contentType, ok := obj["content_type"].(string); ok {
				switch contentType {
				case "stream":
					return streamType
				case "movie":
					return movieType
				case "series":
					return seriesType
				}
			}
		}
		return streamType
	},
})

// =====================================================
// QUERY DEFINITIONS
// =====================================================

func (s *GraphQLService) buildSchema() (graphql.Schema, error) {
	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			// User queries
			"me": &graphql.Field{
				Type: userType,
				Resolve: s.resolveCurrentUser,
			},

			// Content queries
			"streams": &graphql.Field{
				Type: graphql.NewList(streamType),
				Args: graphql.FieldConfigArgument{
					"category": &graphql.ArgumentConfig{Type: graphql.String},
					"limit":    &graphql.ArgumentConfig{Type: graphql.Int},
					"offset":   &graphql.ArgumentConfig{Type: graphql.Int},
				},
				Resolve: s.resolveStreams,
			},

			"stream": &graphql.Field{
				Type: streamType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: s.resolveStream,
			},

			"movies": &graphql.Field{
				Type: graphql.NewList(movieType),
				Args: graphql.FieldConfigArgument{
					"genre":  &graphql.ArgumentConfig{Type: graphql.String},
					"limit":  &graphql.ArgumentConfig{Type: graphql.Int},
					"offset": &graphql.ArgumentConfig{Type: graphql.Int},
				},
				Resolve: s.resolveMovies,
			},

			"movie": &graphql.Field{
				Type: movieType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: s.resolveMovie,
			},

			"series": &graphql.Field{
				Type: graphql.NewList(seriesType),
				Args: graphql.FieldConfigArgument{
					"genre":  &graphql.ArgumentConfig{Type: graphql.String},
					"limit":  &graphql.ArgumentConfig{Type: graphql.Int},
					"offset": &graphql.ArgumentConfig{Type: graphql.Int},
				},
				Resolve: s.resolveSeries,
			},

			"categories": &graphql.Field{
				Type:    graphql.NewList(categoryType),
				Resolve: s.resolveCategories,
			},

			// Recommendations
			"recommendations": &graphql.Field{
				Type: graphql.NewList(recommendationType),
				Args: graphql.FieldConfigArgument{
					"limit": &graphql.ArgumentConfig{Type: graphql.Int},
				},
				Resolve: s.resolveRecommendations,
			},

			"trending": &graphql.Field{
				Type: graphql.NewList(contentType),
				Args: graphql.FieldConfigArgument{
					"limit": &graphql.ArgumentConfig{Type: graphql.Int},
				},
				Resolve: s.resolveTrending,
			},

			// Search
			"search": &graphql.Field{
				Type: graphql.NewList(contentType),
				Args: graphql.FieldConfigArgument{
					"query": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"type":  &graphql.ArgumentConfig{Type: graphql.String},
					"limit": &graphql.ArgumentConfig{Type: graphql.Int},
				},
				Resolve: s.resolveSearch,
			},
		},
	})

	// Mutations
	rootMutation := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"addToWatchlist": &graphql.Field{
				Type: graphql.Boolean,
				Args: graphql.FieldConfigArgument{
					"contentId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"type":      &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: s.resolveAddToWatchlist,
			},

			"removeFromWatchlist": &graphql.Field{
				Type: graphql.Boolean,
				Args: graphql.FieldConfigArgument{
					"contentId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: s.resolveRemoveFromWatchlist,
			},

			"addToFavorites": &graphql.Field{
				Type: graphql.Boolean,
				Args: graphql.FieldConfigArgument{
					"contentId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"type":      &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: s.resolveAddToFavorites,
			},

			"trackView": &graphql.Field{
				Type: graphql.Boolean,
				Args: graphql.FieldConfigArgument{
					"contentId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"type":      &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"position":  &graphql.ArgumentConfig{Type: graphql.Int},
					"duration":  &graphql.ArgumentConfig{Type: graphql.Int},
				},
				Resolve: s.resolveTrackView,
			},
		},
	})

	return graphql.NewSchema(graphql.SchemaConfig{
		Query:    rootQuery,
		Mutation: rootMutation,
	})
}

// =====================================================
// RESOLVERS
// =====================================================

func (s *GraphQLService) resolveCurrentUser(p graphql.ResolveParams) (interface{}, error) {
	userID := p.Context.Value("user_id").(string)
	if userID == "" {
		return nil, errors.New("not authenticated")
	}

	var user map[string]interface{}
	query := `
		SELECT u.id, u.username, u.email, u.avatar_url, p.name as package_name, u.is_active, u.created_at
		FROM users u
		LEFT JOIN packages p ON p.id = u.package_id
		WHERE u.id = ?
	`
	row := s.db.QueryRowContext(p.Context, query, userID)

	var id, username, email, avatarURL, packageName string
	var isActive bool
	var createdAt time.Time

	err := row.Scan(&id, &username, &email, &avatarURL, &packageName, &isActive, &createdAt)
	if err != nil {
		return nil, err
	}

	user = map[string]interface{}{
		"id":           id,
		"username":     username,
		"email":        email,
		"avatar_url":   avatarURL,
		"package_name": packageName,
		"is_active":    isActive,
		"created_at":   createdAt,
	}

	return user, nil
}

func (s *GraphQLService) resolveStreams(p graphql.ResolveParams) (interface{}, error) {
	category, _ := p.Args["category"].(string)
	limit, _ := p.Args["limit"].(int)
	offset, _ := p.Args["offset"].(int)

	if limit == 0 {
		limit = 20
	}

	query := `
		SELECT s.id, s.name as title, s.description, s.poster_url, s.rating,
		       s.is_live, s.created_at, c.name as category_name
		FROM streams s
		LEFT JOIN categories c ON c.id = s.category_id
		WHERE s.is_active = 1
	`

	args := []interface{}{}
	if category != "" {
		query += " AND c.name = ?"
		args = append(args, category)
	}

	query += " ORDER BY s.created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(p.Context, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var streams []map[string]interface{}
	for rows.Next() {
		var id, title, description, posterURL, categoryName string
		var rating float64
		var isLive bool
		var createdAt time.Time

		err := rows.Scan(&id, &title, &description, &posterURL, &rating, &isLive, &createdAt, &categoryName)
		if err != nil {
			continue
		}

		streams = append(streams, map[string]interface{}{
			"id":           id,
			"title":        title,
			"description":  description,
			"poster_url":   posterURL,
			"rating":       rating,
			"is_live":      isLive,
			"created_at":   createdAt,
			"content_type": "stream",
		})
	}

	return streams, nil
}

func (s *GraphQLService) resolveStream(p graphql.ResolveParams) (interface{}, error) {
	id := p.Args["id"].(string)

	var stream map[string]interface{}
	query := `
		SELECT s.id, s.name as title, s.description, s.poster_url, s.rating, s.stream_url, s.is_live
		FROM streams s
		WHERE s.id = ? AND s.is_active = 1
	`

	var title, description, posterURL, streamURL string
	var rating float64
	var isLive bool

	err := s.db.QueryRowContext(p.Context, query, id).Scan(&id, &title, &description, &posterURL, &rating, &streamURL, &isLive)
	if err != nil {
		return nil, err
	}

	stream = map[string]interface{}{
		"id":           id,
		"title":        title,
		"description":  description,
		"poster_url":   posterURL,
		"rating":       rating,
		"stream_url":   streamURL,
		"is_live":      isLive,
		"content_type": "stream",
	}

	return stream, nil
}

func (s *GraphQLService) resolveMovies(p graphql.ResolveParams) (interface{}, error) {
	limit, _ := p.Args["limit"].(int)
	if limit == 0 {
		limit = 20
	}

	query := `
		SELECT id, title, description, poster_url, rating, duration, year, created_at
		FROM vod_movies
		WHERE is_active = 1
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := s.db.QueryContext(p.Context, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []map[string]interface{}
	for rows.Next() {
		var id, title, description, posterURL string
		var rating float64
		var duration, year int
		var createdAt time.Time

		err := rows.Scan(&id, &title, &description, &posterURL, &rating, &duration, &year, &createdAt)
		if err != nil {
			continue
		}

		movies = append(movies, map[string]interface{}{
			"id":           id,
			"title":        title,
			"description":  description,
			"poster_url":   posterURL,
			"rating":       rating,
			"duration":     duration,
			"year":         year,
			"created_at":   createdAt,
			"content_type": "movie",
		})
	}

	return movies, nil
}

func (s *GraphQLService) resolveMovie(p graphql.ResolveParams) (interface{}, error) {
	id := p.Args["id"].(string)

	query := `
		SELECT id, title, description, poster_url, rating, duration, year, director
		FROM vod_movies
		WHERE id = ? AND is_active = 1
	`

	var title, description, posterURL, director string
	var rating float64
	var duration, year int

	err := s.db.QueryRowContext(p.Context, query, id).Scan(&id, &title, &description, &posterURL, &rating, &duration, &year, &director)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":           id,
		"title":        title,
		"description":  description,
		"poster_url":   posterURL,
		"rating":       rating,
		"duration":     duration,
		"year":         year,
		"director":     director,
		"content_type": "movie",
	}, nil
}

func (s *GraphQLService) resolveSeries(p graphql.ResolveParams) (interface{}, error) {
	limit, _ := p.Args["limit"].(int)
	if limit == 0 {
		limit = 20
	}

	query := `SELECT id, title, description, poster_url, rating FROM vod_series WHERE is_active = 1 LIMIT ?`
	rows, err := s.db.QueryContext(p.Context, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var series []map[string]interface{}
	for rows.Next() {
		var id, title, description, posterURL string
		var rating float64

		rows.Scan(&id, &title, &description, &posterURL, &rating)
		series = append(series, map[string]interface{}{
			"id":           id,
			"title":        title,
			"description":  description,
			"poster_url":   posterURL,
			"rating":       rating,
			"content_type": "series",
		})
	}

	return series, nil
}

func (s *GraphQLService) resolveCategories(p graphql.ResolveParams) (interface{}, error) {
	query := `SELECT id, name, description, icon, display_order FROM categories WHERE is_active = 1 ORDER BY display_order`
	rows, err := s.db.QueryContext(p.Context, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []map[string]interface{}
	for rows.Next() {
		var id, name, description, icon string
		var order int
		rows.Scan(&id, &name, &description, &icon, &order)
		categories = append(categories, map[string]interface{}{
			"id":          id,
			"name":        name,
			"description": description,
			"icon":        icon,
			"order":       order,
		})
	}

	return categories, nil
}

func (s *GraphQLService) resolveRecommendations(p graphql.ResolveParams) (interface{}, error) {
	userID := p.Context.Value("user_id").(string)
	limit, _ := p.Args["limit"].(int)
	if limit == 0 {
		limit = 10
	}

	// Call AI recommendation service (simplified)
	query := `
		SELECT content_id, content_type, reason, score
		FROM ai_recommendations
		WHERE user_id = ?
		ORDER BY score DESC
		LIMIT ?
	`
	rows, err := s.db.QueryContext(p.Context, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recommendations []map[string]interface{}
	for rows.Next() {
		var contentID, contentType, reason string
		var score float64
		rows.Scan(&contentID, &contentType, &reason, &score)

		recommendations = append(recommendations, map[string]interface{}{
			"content_id": contentID,
			"type":       contentType,
			"reason":     reason,
			"score":      score,
		})
	}

	return recommendations, nil
}

func (s *GraphQLService) resolveTrending(p graphql.ResolveParams) (interface{}, error) {
	limit, _ := p.Args["limit"].(int)
	if limit == 0 {
		limit = 10
	}

	query := `
		SELECT id, title, poster_url, 'stream' as content_type
		FROM streams
		ORDER BY views DESC
		LIMIT ?
	`
	rows, err := s.db.QueryContext(p.Context, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trending []map[string]interface{}
	for rows.Next() {
		var id, title, posterURL, contentType string
		rows.Scan(&id, &title, &posterURL, &contentType)
		trending = append(trending, map[string]interface{}{
			"id":           id,
			"title":        title,
			"poster_url":   posterURL,
			"content_type": contentType,
		})
	}

	return trending, nil
}

func (s *GraphQLService) resolveSearch(p graphql.ResolveParams) (interface{}, error) {
	query := p.Args["query"].(string)
	searchType, _ := p.Args["type"].(string)
	limit, _ := p.Args["limit"].(int)
	if limit == 0 {
		limit = 20
	}

	searchQuery := fmt.Sprintf("%%%s%%", query)

	sqlQuery := `
		SELECT id, name as title, poster_url, 'stream' as content_type
		FROM streams
		WHERE name LIKE ?
		LIMIT ?
	`

	rows, err := s.db.QueryContext(p.Context, sqlQuery, searchQuery, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id, title, posterURL, contentType string
		rows.Scan(&id, &title, &posterURL, &contentType)
		results = append(results, map[string]interface{}{
			"id":           id,
			"title":        title,
			"poster_url":   posterURL,
			"content_type": contentType,
		})
	}

	return results, nil
}

func (s *GraphQLService) resolveAddToWatchlist(p graphql.ResolveParams) (interface{}, error) {
	userID := p.Context.Value("user_id").(string)
	contentID := p.Args["contentId"].(string)
	contentType := p.Args["type"].(string)

	query := `INSERT INTO user_watchlist (user_id, content_id, content_type, added_at) VALUES (?, ?, ?, NOW())`
	_, err := s.db.ExecContext(p.Context, query, userID, contentID, contentType)
	return err == nil, err
}

func (s *GraphQLService) resolveRemoveFromWatchlist(p graphql.ResolveParams) (interface{}, error) {
	userID := p.Context.Value("user_id").(string)
	contentID := p.Args["contentId"].(string)

	query := `DELETE FROM user_watchlist WHERE user_id = ? AND content_id = ?`
	_, err := s.db.ExecContext(p.Context, query, userID, contentID)
	return err == nil, err
}

func (s *GraphQLService) resolveAddToFavorites(p graphql.ResolveParams) (interface{}, error) {
	userID := p.Context.Value("user_id").(string)
	contentID := p.Args["contentId"].(string)
	contentType := p.Args["type"].(string)

	query := `INSERT INTO user_favorites (user_id, content_id, content_type, added_at) VALUES (?, ?, ?, NOW())`
	_, err := s.db.ExecContext(p.Context, query, userID, contentID, contentType)
	return err == nil, err
}

func (s *GraphQLService) resolveTrackView(p graphql.ResolveParams) (interface{}, error) {
	userID := p.Context.Value("user_id").(string)
	contentID := p.Args["contentId"].(string)
	contentType := p.Args["type"].(string)
	position, _ := p.Args["position"].(int)
	duration, _ := p.Args["duration"].(int)

	query := `
		INSERT INTO view_history (user_id, content_id, content_type, position, duration, viewed_at)
		VALUES (?, ?, ?, ?, ?, NOW())
		ON DUPLICATE KEY UPDATE position = ?, viewed_at = NOW()
	`
	_, err := s.db.ExecContext(p.Context, query, userID, contentID, contentType, position, duration, position)
	return err == nil, err
}

// ExecuteQuery executes a GraphQL query
func (s *GraphQLService) ExecuteQuery(ctx context.Context, query string, variables map[string]interface{}) *graphql.Result {
	params := graphql.Params{
		Schema:         s.schema,
		RequestString:  query,
		VariableValues: variables,
		Context:        ctx,
	}
	return graphql.Do(params)
}
