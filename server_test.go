package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

func TestUnixServerDoubleClose(t *testing.T) {
	server, err := NewUnixServer(&NilMonitor{}, &EchoHandler{}, "/tmp/maxwells-daemon.sock")
	if err != nil {
		t.Fatalf("error starting server: %v", err)
	}
	err = server.Close()
	if err != nil {
		t.Fatalf("error closing server: %v", err)
	}
	err = server.Close()
	if err == nil {
		t.Fatalf("server was able to be doubly-closed")
	}
}

func TestUnixServerInvalidFilename(t *testing.T) {
	server, err := NewUnixServer(&NilMonitor{}, &EchoHandler{}, "/this/directory/doesnt.exist")
	if err == nil {
		server.Close()
		t.Fatalf("no error on invalid filename")
	}
}

func TestUnixServing(t *testing.T) {
	server, err := NewUnixServer(&NilMonitor{}, &EchoHandler{}, "/tmp/maxwells-daemon.sock")
	if err != nil {
		t.Fatalf("error starting server: %v", err)
	}
	defer server.Close()
	connection, err := net.Dial("unix", "/tmp/maxwells-daemon.sock")
	if err != nil {
		t.Fatalf("error connecting to server: %v", err)
	}
	connection.SetDeadline(time.Now().Add(time.Millisecond))
	sample := "0.123456789\n"
	fmt.Fprintf(connection, sample)
	response, err := bufio.NewReader(connection).ReadString('\n')
	if err != nil {
		t.Fatalf("could not read result from server: %v", err)
	}
	if sample != response {
		t.Fatalf("sample '%v' does not match response '%v'", sample, response)
	}
}

const ttcInput string = `
Chapter 1

The Tao that can be spoken is not the eternal Tao
The name that can be named is not the eternal name
The nameless is the origin of Heaven and Earth
The named is the mother of myriad things
Thus, constantly without desire, one observes its essence
Constantly with desire, one observes its manifestations
These two emerge together but differ in name
The unity is said to be the mystery
Mystery of mysteries, the door to all wonders


Chapter 2

When the world knows beauty as beauty, ugliness arises
When it knows good as good, evil arises
Thus being and non-being produce each other
Difficult and easy bring about each other
Long and short reveal each other
High and low support each other
Music and voice harmonize each other
Front and back follow each other
Therefore the sages:
Manage the work of detached actions
Conduct the teaching of no words
They work with myriad things but do not control
They create but do not possess
They act but do not presume
They succeed but do not dwell on success
It is because they do not dwell on success
That it never goes away


Chapter 3

Do not glorify the achievers
So the people will not squabble
Do not treasure goods that are hard to obtain
So the people will not become thieves
Do not show the desired things
So their hearts will not be confused

Thus the governance of the sage:
Empties their hearts
Fills their bellies
Weakens their ambitions
Strengthens their bones

Let the people have no cunning and no greed
So those who scheme will not dare to meddle

Act without contrivance
And nothing will be beyond control


Chapter 4

The Tao is empty
When utilized, it is not filled up
So deep! It seems to be the source of all things

It blunts the sharpness
Unravels the knots
Dims the glare
Mixes the dusts

So indistinct! It seems to exist
I do not know whose offspring it is
Its image is the predecessor of the Emperor


Chapter 5

Heaven and Earth are impartial
They regard myriad things as straw dogs
The sages are impartial
They regard people as straw dogs

The space between Heaven and Earth
Is it not like a bellows?
Empty, and yet never exhausted
It moves, and produces more

Too many words hasten failure
Cannot compare to keeping to the void


Chapter 6

The valley spirit, undying
Is called the Mystic Female

The gate of the Mystic Female
Is called the root of Heaven and Earth

It flows continuously, barely perceptible
Utilize it; it is never exhausted


Chapter 7

Heaven and Earth are everlasting
The reason Heaven and Earth can last forever
Is that they do not exist for themselves
Thus they can last forever

Therefore the sages:
Place themselves last but end up in front
Are outside of themselves and yet survive
Is it not due to their selflessness?
That is how they can achieve their own goals


Chapter 8

The highest goodness resembles water
Water greatly benefits myriad things without contention
It stays in places that people dislike
Therefore it is similar to the Tao

Dwelling with the right location
Feeling with great depth
Giving with great kindness
Speaking with great integrity
Governing with great administration
Handling with great capability
Moving with great timing

Because it does not contend
It is therefore beyond reproach


Chapter 9

Holding a cup and overfilling it
Cannot be as good as stopping short
Pounding a blade and sharpening it
Cannot be kept for long

Gold and jade fill up the room
No one is able to protect them
Wealth and position bring arrogance
And leave disasters upon oneself

When achievement is completed, fame is attained
Withdraw oneself
This is the Tao of Heaven


Chapter 10

In holding the soul and embracing oneness
Can one be steadfast, without straying?
In concentrating the energy and reaching relaxation
Can one be like an infant?
In cleaning away the worldly view
Can one be without imperfections?
In loving the people and ruling the nation
Can one be without manipulation?
In the heavenly gate's opening and closing
Can one hold to the feminine principle?
In understanding clearly all directions
Can one be without intellectuality?

Bearing it, rearing it
Bearing without possession
Achieving without arrogance
Raising without domination
This is called the Mystic Virtue


Chapter 11

Thirty spokes join in one hub
In its emptiness, there is the function of a vehicle
Mix clay to create a container
In its emptiness, there is the function of a container
Cut open doors and windows to create a room
In its emptiness, there is the function of a room

Therefore, that which exists is used to create benefit
That which is empty is used to create functionality


Chapter 12

The five colors make one blind in the eyes
The five sounds make one deaf in the ears
The five flavors make one tasteless in the mouth

Racing and hunting make one wild in the heart
Goods that are difficult to acquire make one cause damage

Therefore the sages care for the stomach and not the eyes
That is why they discard the other and take this


Chapter 13

Favor and disgrace make one fearful
The greatest misfortune is the self
What does "favor and disgrace make one fearful" mean?
Favor is high; disgrace is low
Having it makes one fearful
Losing it makes one fearful
This is "favor and disgrace make one fearful"

What does "the greatest misfortune is the self" mean?
The reason I have great misfortune
Is that I have the self
If I have no self
What misfortune do I have?

So one who values the self as the world
Can be given the world
One who loves the self as the world
Can be entrusted with the world


Chapter 14

Look at it, it cannot be seen
It is called colorless
Listen to it, it cannot be heard
It is called noiseless
Reach for it, it cannot be held
It is called formless
These three cannot be completely unraveled
So they are combined into one

Above it, not bright
Below it, not dark
Continuing endlessly, cannot be named
It returns back into nothingness
Thus it is called the form of the formless
The image of the imageless
This is called enigmatic
Confront it, its front cannot be seen
Follow it, its back cannot be seen

Wield the Tao of the ancients
To manage the existence of today
One can know the ancient beginning
It is called the Tao Axiom


Chapter 15

The Tao masters of antiquity
Subtle wonders through mystery
Depths that cannot be discerned
Because one cannot discern them
Therefore one is forced to describe the appearance

Hesitant, like crossing a wintry river
Cautious, like fearing four neighbors
Solemn, like a guest
Loose, like ice about to melt
Genuine, like plain wood
Open, like a valley
Opaque, like muddy water

Who can be muddled yet desist
In stillness gradually become clear?
Who can be serene yet persist
In motion gradually come alive?

One who holds this Tao does not wish to be overfilled
Because one is not overfilled
Therefore one can preserve and not create anew


Chapter 16

Attain the ultimate emptiness
Hold on to the truest tranquility
The myriad things are all active
I therefore watch their return

Everything flourishes; each returns to its root
Returning to the root is called tranquility
Tranquility is called returning to one's nature
Returning to one's nature is called constancy
Knowing constancy is called clarity

Not knowing constancy, one recklessly causes trouble
Knowing constancy is acceptance
Acceptance is impartiality
Impartiality is sovereign
Sovereign is Heaven
Heaven is Tao
Tao is eternal
The self is no more, without danger


Chapter 17

The highest rulers, people do not know they have them
The next level, people love them and praise them
The next level, people fear them
The next level, people despise them
If the rulers' trust is insufficient
Have no trust in them

Proceeding calmly, valuing their words
Task accomplished, matter settled
The people all say, "We did it naturally"


Chapter 18

The great Tao fades away
There is benevolence and justice
Intelligence comes forth
There is great deception

The six relations are not harmonious
There is filial piety and kind affection
The country is in confused chaos
There are loyal ministers


Chapter 19

End sagacity; abandon knowledge
The people benefit a hundred times

End benevolence; abandon righteousness
The people return to piety and charity

End cunning; discard profit
Bandits and thieves no longer exist

These three things are superficial and insufficient
Thus this teaching has its place:
Show plainness; hold simplicity
Reduce selfishness; decrease desires


Chapter 20

Cease learning, no more worries
Respectful response and scornful response
How much is the difference?
Goodness and evil
How much do they differ?
What the people fear, I cannot be unafraid

So desolate! How limitless it is!
The people are excited
As if enjoying a great feast
As if climbing up to the terrace in spring
I alone am quiet and uninvolved
Like an infant not yet smiling
So weary, like having no place to return
The people all have surplus
While I alone seem lacking
I have the heart of a fool indeed – so ignorant!
Ordinary people are bright
I alone am muddled
Ordinary people are scrutinizing
I alone am obtuse
Such tranquility, like the ocean
Such high wind, as if without limits

The people all have goals
And I alone am stubborn and lowly
I alone am different from them
And value the nourishing mother


Chapter 21

The appearance of great virtue
Follows only the Tao
The Tao, as a thing
Seems indistinct, seems unclear

So unclear, so indistinct
Within it there is image
So indistinct, so unclear
Within it there is substance
So deep, so profound
Within it there is essence

Its essence is supremely real
Within it there is faith
From ancient times to the present
Its name never departs
To observe the source of all things
How do I know the nature of the source?
With this


Chapter 22

Yield and remain whole
Bend and remain straight
Be low and become filled
Be worn out and become renewed
Have little and receive
Have much and be confused
Therefore the sages hold to the one as an example for the world
Without flaunting themselves – and so are seen clearly
Without presuming themselves – and so are distinguished
Without praising themselves – and so have merit
Without boasting about themselves – and so are lasting

Because they do not contend, the world cannot contend with them
What the ancients called "the one who yields and remains whole"
Were they speaking empty words?
Sincerity becoming whole, and returning to oneself


Chapter 23

Sparse speech is natural
Thus strong wind does not last all morning
Sudden rain does not last all day
What makes this so? Heaven and Earth
Even Heaven and Earth cannot make it last
How can humans?

Thus those who follow the Tao are with the Tao
Those who follow virtue are with virtue
Those who follow loss are with loss
Those who are with the Tao, the Tao is also pleased to have them
Those who are with virtue, virtue is also pleased to have them
Those who are with loss, loss is also please to have them
Those who do not trust sufficiently, others have no trust in them


Chapter 24

Those who are on tiptoes cannot stand
Those who straddle cannot walk
Those who flaunt themselves are not clear
Those who presume themselves are not distinguished
Those who praise themselves have no merit
Those who boast about themselves do not last

Those with the Tao call such things leftover food or tumors
They despise them
Thus, those who possesses the Tao do not engage in them


Chapter 25

There is something formlessly created
Born before Heaven and Earth
So silent! So ethereal!
Independent and changeless
Circulating and ceaseless
It can be regarded as the mother of the world

I do not know its name
Identifying it, I call it "Tao"
Forced to describe it, I call it great
Great means passing
Passing means receding
Receding means returning
Therefore the Tao is great
Heaven is great
Earth is great
The sovereign is also great
There are four greats in the universe
And the sovereign occupies one of them
Humans follow the laws of Earth
Earth follows the laws of Heaven
Heaven follows the laws of Tao
Tao follows the laws of nature


Chapter 26

Heaviness is the root of lightness
Quietness is the master of restlessness

Therefore the sages travel an entire day
Without leaving the heavy supplies
Even though there are luxurious sights
They are composed and transcend beyond

How can the lords of ten thousand chariots
Apply themselves lightly to the world?
To be light is to lose one's root
To be restless is to lose one's mastery


Chapter 27

Good traveling does not leave tracks
Good speech does not seek faults
Good reckoning does not use counters
Good closure needs no bar and yet cannot be opened
Good knot needs no rope and yet cannot be untied

Therefore sages often save others
And so do not abandon anyone
They often save things
And so do not abandon anything
This is called following enlightenment

Therefore the good person is the teacher of the bad person
The bad person is the resource of the good person
Those who do not value their teachers
And do not love their resources
Although intelligent, they are greatly confused
This is called the essential wonder


Chapter 28

Know the masculine, hold to the feminine
Be the watercourse of the world
Being the watercourse of the world
The eternal virtue does not depart
Return to the state of the infant
Know the white, hold to the black
Be the standard of the world
Being the standard of the world
The eternal virtue does not deviate
Return to the state of the boundless
Know the honor, hold to the humility
Be the valley of the world
Being the valley of the world
The eternal virtue shall be sufficient
Return to the state of plain wood
Plain wood splits, then becomes tools
The sages utilize them
And then become leaders
Thus the greater whole is undivided


Chapter 29

Those who wish to take the world and control it
I see that they cannot succeed
The world is a sacred instrument
One cannot control it
The one who controls it will fail
The one who grasps it will lose

Because all things:
Either lead or follow
Either blow hot or cold
Either have strength or weakness
Either have ownership or take by force

Therefore the sage:
Eliminates extremes
Eliminates excess
Eliminates arrogance


Chapter 30

The one who uses the Tao to advise the ruler
Does not dominate the world with soldiers
Such methods tend to be returned

The place where the troops camp
Thistles and thorns grow
Following the great army
There must be an inauspicious year

A good commander achieves result, then stops
And does not dare to reach for domination
Achieves result but does not brag
Achieves result but does not flaunt
Achieves result but is not arrogant
Achieves result but only out of necessity
Achieves result but does not dominate

Things become strong and then get old
This is called contrary to the Tao
That which is contrary to the Tao soon ends


Chapter 31

A strong military, a tool of misfortune
All things detest it
Therefore, those who possess the Tao avoid it
Honorable gentlemen, while at home, value the left
When deploying the military, value the right

The military is a tool of misfortune
Not the tool of honorable gentlemen
When using it out of necessity
Calm detachment should be above all
Victorious but without glory
Those who glorify
Are delighting in the killing
Those who delight in killing
Cannot achieve their ambitions upon the world

Auspicious events favor the left
Inauspicious events favor the right
The lieutenant general is positioned to the left
The major general is positioned to the right
We say that they are treated as if in a funeral
Those who have been killed
Should be mourned with sadness
Victory in war should be treated as a funeral


Chapter 32

The Tao, eternally nameless
Its simplicity, although imperceptible
Cannot be treated by the world as subservient

If the sovereign can hold on to it
All will follow by themselves
Heaven and Earth, together in harmony
Will rain sweet dew
People will not need to force it; it will adjust by itself

In the beginning, there were names
Names came to exist everywhere
One should know when to stop
Knowing when to stop, thus avoiding danger

The existence of the Tao in the world
Is like streams in the valley into rivers and the ocean


Chapter 33

Those who understand others are intelligent
Those who understand themselves are enlightened

Those who overcome others have strength
Those who overcome themselves are powerful

Those who know contentment are wealthy
Those who proceed vigorously have willpower

Those who do not lose their base endure
Those who die but do not perish have longevity


Chapter 34

The great Tao is like a flood
It can flow to the left or to the right

The myriad things depend on it for life, but it never stops
It achieves its work, but does not take credit
It clothes and feeds myriad things, but does not rule over them

Ever desiring nothing
It can be named insignificant
Myriad things return to it but it does not rule over them
It can be named great

Even in the end, it does not regard itself as great
That is how it can achieve its greatness


Chapter 35

Hold the great image
All under heaven will come
They come without harm, in harmonious peace

Music and food, passing travelers stop
The Tao that is spoken out of the mouth
Is bland and without flavor

Look at it, it cannot be seen
Listen to it, it cannot be heard
Use it, it cannot be exhausted


Chapter 36

If one wishes to shrink it
One must first expand it
If one wishes to weaken it
One must first strengthen it
If one wishes to discard it
One must first promote it
If one wishes to seize it
One must first give it
This is called subtle clarity

The soft and weak overcomes the tough and strong
Fish cannot leave the depths
The sharp instruments of the state
Cannot be shown to the people


Chapter 37

The Tao is constant in non-action
Yet there is nothing it does not do

If the sovereign can hold on to this
All things shall transform themselves
Transformed, yet wishing to achieve
I shall restrain them with the simplicity of the nameless
The simplicity of the nameless
They shall be without desire
Without desire, using stillness
The world shall steady itself
`

func TestUnixServingLargeInput(t *testing.T) {
	server, err := NewUnixServer(&NilMonitor{}, &EchoHandler{}, "/tmp/maxwells-daemon.sock")
	if err != nil {
		t.Fatalf("error starting server: %v", err)
	}
	defer server.Close()
	connection, err := net.Dial("unix", "/tmp/maxwells-daemon.sock")
	if err != nil {
		t.Fatalf("error connecting to server: %v", err)
	}
	connection.SetDeadline(time.Now().Add(time.Millisecond))
	input := strings.Replace(ttcInput, "\n", "\t", -1) + "\n"
	count, err := fmt.Fprintf(connection, input)
	if err != nil {
		t.Fatalf("could not write long text to server: %v", err)
	}
	if count != len(input) {
		t.Fatalf("only wrote %v out of %v bytes to server")
	}
	response, err := bufio.NewReader(connection).ReadString('\n')
	if err != nil {
		t.Fatalf("could not read result from server: %v", err)
	}
	if input != response {
		t.Fatalf("sample '%v' does not match response '%v'", input, response)
	}
}
