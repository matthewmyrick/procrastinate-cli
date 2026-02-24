# Procrastinate TUI

## Procrastinate service

We use postgres with Procrastinate and python to run our queueing system and such

there is no dashboard or any good tools out there to see what is wrong with the queue or what we are doing with it

We will configure to the database based on a config file with the pass and username
This config will also have what defauilt queue, polliing in seconds like how frequent we are polling, etc. 

for now everythign is read commands nothing else

## TUI

The tui is going to be queue based with many panes to show us what we need to do. Also put at the top right what user we are in. We are aloud to change the user as well. So we can have n number of users

so i can change qeues and the panes will be the same. We will show what queue we are on at the top left as well. 

the panes i want 

live tail of the jobs coming in based on that queue
are they getting blocked by anything? sometimes we are getting 

## Code

Please use golang and bubble tea (charm) to create the tui 

then pleae use cobra for the cli commands (we will eventuialy have this biut for now just focus on the )

make sure our code is broken down into smaller files. I dont want code that is like 500 lines of code (only do this if we hacve to). Most files should be sub 400 lines please.
Make sure naming of the files are easy and follow of what the code actually does. 

create two spoerate folders one for the cli and one for tui please

## what to do next

Ask any qwuestions or concerns on this please and create a plan for us to rip! 
