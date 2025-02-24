# Stats.fm Booster

A simple Go program to generate mock Spotify streaming data for importing into StatsFM. Simulate streams for your favorite tracks.

---

## What Does This Do?

This program generates fake Spotify streaming data in JSON format, which you can import into StatsFM. It allows you to simulate streams for any track, album, or artist by providing a Spotify ID or link.

**Note**: StatsFM does not currently verify the legitimacy of imported data, but excessive use of this tool may result in a ban. Use responsibly!

---

## How to Use

### Prerequisites
- Go installed on your machine. Download it [here](https://golang.org/dl/).

### Steps

1. **Clone the Repository**:
   ```bash
   git clone https://github.com/ProbTom/stats.fm-booster
   cd stats.fm-booster
   
Run the Program:
```go run stats.go```
Follow the Prompts:

Enter a Spotify track.

Specify the total number of streams.

Enter a start date and end date for the streams.

Choose whether to customize the output file name (optional).

After running the program, a JSON file (e.g., output.json) will be created in the same folder.

Go to [StatsFM Imports.](https://stats.fm/settings/imports)

Drag and drop the generated JSON file into the import section.

Wait for the data to appear on your profile (usually 1-3 minutes).

---

### Features

Simulate Streams: Generate fake streaming data for any track.

Customizable Output: Optionally customize the output file name.

Flexible Input: Accepts Spotify IDs or links.

Randomized Data: Streams are generated with random timestamps and playback durations.

---

# Warnings

You can get banned if you use this excessively. StatsFM reserves the right to ban your account if they detect suspicious activity.

You need stats.fm Premium.

This program is for educational purposes only. Use it at your own risk.
---

# Author
# [Tom](https://github.com/ProbTom)

### Credits 
### [scoobyluvs](https://github.com/scoobyluvs/StatsFM-Cheat/)

---

### Contact

[Discord](https://discord.com/users/229396464848076800)

Telegram [@ProbTom](https://t.me/ProbTom) 
