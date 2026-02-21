# THIS TOOL WILL NOT WORK ANYMORE DUE TO [SPOTIFY API NEW UPDAPTE](https://developer.spotify.com/blog/2026-02-06-update-on-developer-access-and-platform-security) (GO FUCK URSELF SPOTIFY) PLEASE ONLY USE https://fuckstats.lol FROM NOW ON.


---

# Stats.fm Booster

A simple Go program to generate mock Spotify streaming data for importing into StatsFM. Simulate streams for your favorite tracks.

---

## What Does This Do?

This program generates fake Spotify streaming data in JSON format, which you can import into StatsFM. It allows you to simulate streams for any track, album, or artist by providing a Spotify ID or link.

**Note**: StatsFM does not currently verify the legitimacy of imported data, but excessive use of this tool may result in a ban. Use responsibly!

---

## How to Use
- website : https://fuckstats.lol

- discord server : https://discord.gg/UGV83QfS5U

- If you lazy to read here a [video tutorial](https://www.youtube.com/watch?v=P2EMltNhxE0&t=18s)

### Prerequisites
- Go installed on your machine. Download it [here](https://golang.org/dl/).

### Steps

1. **Clone the Repository**:
   ```bash
   git clone https://github.com/ProbTom/stats.fm-booster
   cd stats.fm-booster
   ```
or download it as zip 

2. **Run the Program**:
   ```bash
   go run stats.go
   ```

3. **Follow the Prompts**:
   - Enter a Spotify track ID or link.
   - Choose whether to maximize streaming density (389,306 streams) or specify a custom number of streams.
   - If not maximizing density:
     - Enter the total number of streams.
     - Specify a start date and end date for the streams.
   - Choose whether to customize the output file name (optional).

4. **Import the Data**:
   - After running the program, a JSON file (e.g., `output.json`) will be created in the same folder.
   - Go to [StatsFM Imports](https://stats.fm/settings/imports).
   - Drag and drop the generated JSON file into the import section.
   - Wait for the data to appear on your profile (usually 2-5 minutes).

---

### Features

- **Simulate Streams**: Generate fake streaming data for any track.
- **Bulk Mode**: Process multiple tracks from a bulk.txt file.
- **Max Streaming Density**: Option to generate **389,306 streams** for maximum density.
- **Customizable Output**: Optionally customize the output file name.
- **Flexible Input**: Accepts Spotify IDs or links.
- **Randomized Data**: Streams are generated with random timestamps and playback durations (unless maximizing density).

---

## Warnings

- **You can get banned if you use this excessively**. StatsFM reserves the right to ban your account if they detect suspicious activity.
- **You need Stats.fm Premium** to use the import feature.
- **This program is for educational purposes only.** Use it at your own risk.

---

## Author

### [Tom](https://github.com/ProbTom)

---

## Credits

### [scoobyluvs](https://github.com/scoobyluvs/StatsFM-Cheat/)

---

## Contact & Support

- [Discord](https://discord.com/users/229396464848076800)
- Telegram [@ProbTom](https://t.me/ProbTom)
- [Donations](https://paypal.me/tomloison)

---


### Infos
- **I do not claim any responsibility if you get banned from the app or leaderboard.**
- **If you encounter any problem hit me up on Discord or Telegram.**
