HOW TO RUN?

go run <filename>

For lamport:  The message should be read according to the sender clock in the message queue. The message queue will appear after the ten iter.

For Vector: when there is "Error!", there are casuality error occurs. The structure of the vector is [client1,client2,client3,...,server].

For Bully: Please refer to the comment. Best condition: highest id machine first find the coordinator is down. Worst condition: the smallest id machine first found coordinator is down.


collaboration: Luo Yifan