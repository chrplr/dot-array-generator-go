dot-array-generator-go
----------------------

This is a port of Lauren S. Aulet's [dot-array-generator](https://github.com/laurenaulet/dot-array-stimulus-toolbox) to [Go](https://go.dev).

Reference:

Aulet, L.S. (2026). Dot Array Stimulus Toolbox: An Open-Source Solution for Generating and Analyzing Non-Symbolic Number Stimuli. PsyArXiv.
 (available [here](https://osf.io/preprints/psyarxiv/uhsv6_v1))

While the original toolbox provides a graphical interface through a browser, this one is only a command line version, with the following parameter:

```
  -avg-radius float
    	Average dot radius in pixels (default 15)
  -control-area
    	Scale dot sizes to reach target cumulative area
  -control-hull
    	Attempt to constrain convex hull area (experimental)
  -count int
    	Number of stimuli to generate (default 10)
  -height int
    	Image height in pixels (default 400)
  -margin int
    	Margin from image edge in pixels (default 20)
  -min-radius float
    	Minimum dot radius in pixels (default 5)
  -min-spacing float
    	Minimum gap between dot edges in pixels (default 2)
  -n int
    	Number of dots (fixed). Ignored if -n-min and -n-max are set differently. (default 20)
  -n-max int
    	Max dots when using a range (0 = use -n)
  -n-min int
    	Min dots when using a range (0 = use -n)
  -no-aa
    	Disable antialiasing
  -output string
    	Output directory for images and ground_truth.csv (default ".")
  -prefix string
    	Filename prefix for generated images (default "stimulus")
  -seed int
    	Random seed (0 = random)
  -size-variability float
    	Size variability: SD of radius (0 = uniform)
  -target-area float
    	Target cumulative area in px² (used with -control-area) (default 5000)
  -target-hull float
    	Target convex hull area in px² (used with -control-hull) (default 50000)
  -white-on-black
    	White dots on black background (default: black on white)
  -width int
    	Image width in pixels (default 400)
```

The port was performed using [Claude Code](https://code.claude.com/docs/en/overview) by [Christophe Pallier](http://www.pallier.org) on March 17, 2026


License: MIT

---

## How to download and run the program (no installation required)

### Step 1 — Download the right file for your computer

Go to the [Releases page](../../releases/latest) and download the file that matches your system:

| Your computer | File to download |
|---|---|
| Windows, standard PC or laptop | `dot-array-generator-Aulet-windows-amd64.exe` |
| Windows, ARM device (e.g. Snapdragon laptop) | `dot-array-generator-Aulet-windows-arm64.exe` |
| Mac with Intel processor (before 2020) | `dot-array-generator-Aulet-darwin-amd64` |
| Mac with Apple Silicon (M1/M2/M3/M4) | `dot-array-generator-Aulet-darwin-arm64` |
| Linux, standard PC or server | `dot-array-generator-Aulet-linux-amd64` |
| Linux, ARM device (e.g. Raspberry Pi) | `dot-array-generator-Aulet-linux-arm64` |

Not sure which Mac you have? Click the Apple menu () → **About This Mac**. If it says "Apple M1" (or M2/M3/M4), choose the Apple Silicon file. If it says "Intel", choose the Intel file.

### Step 2 — Prepare the file

**On Windows:** No extra step needed — the `.exe` file is ready to use.

**On Mac or Linux:** You need to make the file executable. Open a Terminal, navigate to the folder where you downloaded the file, and run:

```
chmod +x dot-array-generator-Aulet-darwin-arm64
```

(replace `darwin-arm64` with the variant you downloaded)

**On Mac only:** The first time you run it, macOS may warn you that the developer is unknown. To allow it:
1. In **Finder**, right-click (or Control-click) the file and choose **Open**.
2. Click **Open** in the dialog that appears.
3. After that, you can run it normally from the Terminal.

### Step 3 — Run the program

Open a terminal (on Windows: **Command Prompt** or **PowerShell**; on Mac/Linux: **Terminal**), navigate to the folder containing the downloaded file, and run it:

**Windows:**
```
dot-array-generator-Aulet-windows-amd64.exe -n 20 -count 10 -output my_stimuli
```

**Mac (Apple Silicon):**
```
./dot-array-generator-Aulet-darwin-arm64 -n 20 -count 10 -output my_stimuli
```

**Linux (amd64):**
```
./dot-array-generator-Aulet-linux-amd64 -n 20 -count 10 -output my_stimuli
```

This example generates 10 images, each containing 20 dots, and saves them in a folder called `my_stimuli`. See the parameter list above for all available options.
