package spotify

import (
	"context"
	"errors"
	"fmt"
	"time"

	"TeleBotNotifications/internal/logger"
)

func (c *Client) GetNewReleasesArtist(artist Artist, token OAuth2Token, rangeStart, rangeEnd time.Time, ctx context.Context) ([]Album, error) {
	var newAlbums []Album
	lastAlbums, err := c.GetArtistAlbums(&token, &artist)
	if err != nil {
		return nil, fmt.Errorf("error getting albums for artist %s(%s): %s", artist.Name, artist.Id, err)
	}
	for _, album := range lastAlbums {
		select {
		case <-ctx.Done():
			return nil, context.Canceled
		default:
			if !rangeStart.After(album.ReleaseDate) && !rangeEnd.Before(album.ReleaseDate) {
				newAlbums = append(newAlbums, album)
			}
		}
	}
	return newAlbums, nil
}

func (c *Client) GetNewReleases(token OAuth2Token, rangeStart, rangeEnd time.Time, ctx context.Context) ([]Album, error) {
	artists, err := c.GetFollowedArtists(&token)
	if err != nil {
		return nil, fmt.Errorf("error getting artists: %w", err)
	}
	logger.General.Println("Going to check", len(artists), "artists")

	var newAlbums []Album
	for _, artist := range artists {
		select {
		case <-ctx.Done():
			return nil, context.Canceled
		default:
			newArtistsAlbums, err := c.GetNewReleasesArtist(artist, token, rangeStart, rangeEnd, ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return nil, err
				}
				continue

			}
			if newArtistsAlbums != nil {
				newAlbums = append(newAlbums, newArtistsAlbums...)
			}
			time.Sleep(1 * time.Second)
		}
	}
	return newAlbums, nil
}
