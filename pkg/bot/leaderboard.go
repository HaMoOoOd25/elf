package bot

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/nint8835/elf/pkg/adventofcode"
	"github.com/nint8835/elf/pkg/database"
)

func (bot *Bot) GenerateLeaderboardEmbed(guildId string) (*discordgo.MessageEmbed, error) {
	var guild database.Guild
	if tx := bot.Database.First(&guild, "guild_id = ?", guildId); tx.Error != nil {
		return nil, fmt.Errorf("error fetching guild details: %w", tx.Error)
	}

	if guild.LeaderboardID == nil {
		return nil, errors.New("no leaderboard id set")
	}

	leaderboard, err := bot.AdventOfCodeClient.GetLeaderboard(*guild.LeaderboardID, bot.Config.AdventOfCodeEvent)
	if err != nil {
		return nil, fmt.Errorf("error fetching leaderboard: %w", err)
	}

	leaderboardEmbed := &discordgo.MessageEmbed{
		Title:  "Leaderboard",
		URL:    fmt.Sprintf("https://adventofcode.com/%s/leaderboard/private/view/%s", bot.Config.AdventOfCodeEvent, *guild.LeaderboardID),
		Color:  0x007152,
		Fields: []*discordgo.MessageEmbedField{},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Join code: %s", *guild.LeaderboardCode),
		},
		Timestamp: leaderboard.RetrievedAt.Format(time.RFC3339),
	}

	leaderboardEntries := []adventofcode.LeaderboardMember{}

	for _, member := range leaderboard.Leaderboard.Members {
		leaderboardEntries = append(leaderboardEntries, member)
	}

	sort.Slice(leaderboardEntries, func(i, j int) bool {
		return leaderboardEntries[i].LocalScore > leaderboardEntries[j].LocalScore
	})

	usernames := ""
	points := ""
	all_stars := ""

	for i, member := range leaderboardEntries[:int(math.Min(float64(len(leaderboardEntries)), 20))] {
		stars := "`"

		for dayNumber := 1; dayNumber <= 25; dayNumber++ {
			day, ok := member.CompletionDayLevel[strconv.Itoa(dayNumber)]
			if !ok {
				stars += " "
				continue
			}
			_, star1 := day["1"]
			_, star2 := day["2"]
			if star1 && star2 {
				stars += "★"
			} else if star1 || star2 {
				stars += "☆"
			} else {
				stars += "▪"
			}
		}

		stars = strings.TrimRight(stars, " ")

		stars += "`"

		username := member.Name
		
		if username == "" {
			username = fmt.Sprintf("(anonymous user #%s)", member.ID)
		}

		usernames += fmt.Sprintf("`%d` %s\n", i+1, username)
		points += fmt.Sprintf("`%d`\n", member.LocalScore)
		all_stars += stars + "\n"
	}

	leaderboardEmbed.Fields = append(leaderboardEmbed.Fields, &discordgo.MessageEmbedField{
		Name:  "Top 20",
		Value: usernames,
		Inline: true,
	})
	leaderboardEmbed.Fields = append(leaderboardEmbed.Fields, &discordgo.MessageEmbedField{
		Name:  "Points",
		Value: points,
		Inline: true,
	})
	leaderboardEmbed.Fields = append(leaderboardEmbed.Fields, &discordgo.MessageEmbedField{
		Name:  "Stars",
		Value: all_stars,
		Inline: true,
	})

	return leaderboardEmbed, nil
}
