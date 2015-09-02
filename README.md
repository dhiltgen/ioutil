# ioutil
CLI tool to display iostat utilization suitable for xmobar


Here's an example ~/.xmobarrc file

```haskell
Config { font = "-*-Fixed-Bold-R-Normal-*-13-*-*-*-*-*-*-*"
       , bgColor = "black"
       , fgColor = "grey"
       , position = TopW L 90
       , commands = [ Run Cpu ["-L","3","-H","50","--normal","green","--high","red"] 10
                    , Run Memory ["-t","Mem: <usedratio>%"] 10
                    , Run Swap [] 10
                    , Run Date "%a %b %_d %l:%M" "date" 10
                    , Run StdinReader
                    , Run CommandReader "/home/daniel/bin/ledmon" "LED"
                    , Run Com "bash" ["-c", "/usr/bin/docker ps -q | wc -l"] "containers" 100
                    , Run Com "/home/daniel/bin/ioutil" ["--interval=5", "--color"] "ioutil" 100
                    , Run BatteryP ["BAT0"]
                        ["-t", "<acstatus> (<left>%)",
                        "-L", "10", "-H", "80", "-p", "3",
                        "--", "-O", "<fc=green>On</fc> - ", "-i", "",
                        "-L", "-15", "-H", "-5",
                        "-l", "red", "-m", "blue", "-h", "green"]
                        600
                    , Run CoreTemp [] 50
                    ]
       , sepChar = "%"
       , alignSep = "}{"
       , template = "%StdinReader% }{ <fc=#ffff00>%LED%</fc> %cpu% %coretemp% | %memory% | %swap% I/O: %ioutil% Containers: %containers% %battery% <fc=#ee9a00>%date%</fc> "
       }

```
