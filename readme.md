<!-- @format -->

# Reticulum: A Two-Layer Blockchain Sharding Protocol





### Commands for configuring the environment：

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





- There are 16 experiments of Reticulum totally with four kind of attacks(average, bankrun, suicidal, worst) and four Pa ratio(10%,20%,30%,33%) 
- 16 experiments are in different folder, such as 'Reti_code_average_10' means the experiment of Reticulum under average attack and Pa=10%.

- Each experiment should run the start.sh in each folder. For example, if you want to run the experiment of Reticulum under average attack and Pa=10%, you should input './start.sh' in /Reti_code_average_10.

- Each experiment will last about an hour and run for about 100 epochs. The terminal will continue to output a number of statements to prove that the experiment is proceeding in an orderly fashion, however, its purpose is to simply prove that the experiment is proceeding and you can do so without having to pay attention to the output (much of which is information that developers need to pay attention to in order to perform debugging). At the end of the experiment the program is automatically stopped.

- The output of each experiment is kept under the corresponding 'data' folder (and you don't need to pay much attention to it, as we will focus on data visualization).


- To make it easier for you to run experiments, you can run all 16 experiments sequentially (i.e. you don't need to click on each folder to run them) via runall.sh. Specifically, in the top folder directory, input '. /runall.sh' in the top folder. If you meet "bash: ./runall.sh: Permission denied", you could use 'chmod +x runall.sh' to solve it and then '. /runall.sh'.



- After all experiments, "Figure of experiment.ipynb" is used for draw the figure and visualize the result.

