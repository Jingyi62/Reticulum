<!-- @format -->

# A Two-Layer Blockchain Sharding Protocol Leveraging Safety and Liveness for Enhanced Performance

The Artifact has two part:

- 1. The experiment of Rapidchain and Reticulum
- 2. The simulation of Gearbox

### 1. The experiment of Rapidchain and Reticulum

#### Commands for configuring the environment：

- If you use AWS as our provided, you don't need to configure the environment. Otherwise, if you use your own machine, please use the following commands to do that.

```chmod +x start.sh
sudo add-apt-repository ppa:longsleep/golang-backports
sudo apt update
sudo apt install golang-1.19
echo 'export PATH="/usr/lib/go-1.19/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
go version
go mod init example.com/m
go mod init example.com/m/v2
go get gopkg.in/yaml.v2
go get -u github.com/stretchr/testify
go get -u github.com/mr-tron/base58
```

#### Notes:

- There are 16 experiments of Reticulum totally with four kind of attacks (average, bankrun, suicidal, worst) and four $P_{a-run}$ ratio (10%,20%,30%,33%)
- 16 experiments are in different folder, such as 'Reti_code_average_10' means the experiment of Reticulum under average attack in $P_{a-run}$ = 10%; "Reti_code_suicidal_33" means the experiment of Recticulum under suicidal attack in $P_{a-run}$ = 33%.

**How to Run?**

- Each experiment should run the **_start.sh_** in each folder. For example, if you want to run the experiment of Reticulum under average attack and $P_a$=10%, you should execute '.**_/start.sh_**' in '/Reti_code_average_10' folder.
- Each experiment will last about an hour and run for about 100 calendar elements. The terminal will continue to output some statements, but its purpose is just to prove that the experiment is going on, and you don't have to pay attention to the output (much of the information in it is something that developers need to pay attention to when they are debugging). At the end of the experiment, the program stops automatically.
- To make it easier for you to run experiments, you can run all 16 experiments sequentially (i.e. you don't need to click on each folder to run them) via **_runall.sh_**. Specifically, execute it in the top folder directory. If you meet "bash: ./runall.sh: Permission denied", you could use 'chmod +x runall.sh' to solve it and then '. /runall.sh'. (But this will cost you more time, you should make sure that your PC keep running until all 16 experiments finished. This is why we highly recommend that you use the AWS that we provided. )
- After all experiments, "Figure of experiment.ipynb" is used for draw the figure. You just need to execute each module in this jupyter notebook sequentially to get the visualization of the experiment.

### 2. The simulation of Gearbox

- You can see the exact simulation steps via 'Gearbox.ipynb'
