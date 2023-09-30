<!-- @format -->

- There are 16 experiments of Reticulum totally with four kind of attacks(average, bankrun, suicidal, worst) and four Pa ratio(10%,20%,30%,33%) 
- 16 experiments are in different folder, such as 'Reti_code_average_10' means the experiment of Reticulum under average attack and Pa=10%.

- Each experiment should run the start.sh in each folder. For example, if you want to run the experiment of Reticulum under average attack and Pa=10%, you should input './start.sh' in /Reti_code_average_10.

- Each experiment will last about an hour and run for about 100 epochs. The terminal will continue to output a number of statements to prove that the experiment is proceeding in an orderly fashion, however, its purpose is to simply prove that the experiment is proceeding and you can do so without having to pay attention to the output (much of which is information that developers need to pay attention to in order to perform debugging). At the end of the experiment the program is automatically stopped.

- The output of each experiment is kept under the corresponding 'data' folder (and you don't need to pay much attention to it, as we will focus on data visualization).


- To make it easier for you to run experiments, you can run all 16 experiments sequentially (i.e. you don't need to click on each folder to run them) via runall.sh. Specifically, in the top folder directory, input '. /runall.sh' in the top folder. If you meet "bash: ./runall.sh: Permission denied", you could use 'chmod +x runall.sh' to solve it and then '. /runall.sh'.

