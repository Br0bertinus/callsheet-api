package service

// dailyActorPool is a curated set of prolific, well-connected TMDB actor IDs.
// Every actor here has a large filmography, making a valid chain between any
// two almost always achievable.
//
// To add more actors: add the name to scripts/rebuild_actor_pool/main.go and
// re-run:  TMDB_API_KEY=<key> go run ./scripts/rebuild_actor_pool
// Find an actor's TMDB ID at https://www.themoviedb.org/person/<id>.
var dailyActorPool = []int{
	// --- A-list headliners ---
	31,    // Tom Hanks
	287,   // Brad Pitt
	380,   // Robert De Niro
	500,   // Tom Cruise
	514,   // Jack Nicholson
	192,   // Morgan Freeman
	1158,  // Al Pacino
	4483,  // Dustin Hoffman
	5292,  // Denzel Washington
	2231,  // Samuel L. Jackson
	6193,  // Leonardo DiCaprio
	1892,  // Matt Damon
	6384,  // Keanu Reeves
	4724,  // Kevin Bacon
	2888,  // Will Smith
	1204,  // Julia Roberts
	5064,  // Meryl Streep
	6968,  // Hugh Jackman
	1245,  // Scarlett Johansson
	524,   // Natalie Portman
	85,    // Johnny Depp
	880,   // Ben Affleck
	62,    // Bruce Willis
	6885,  // Charlize Theron
	934,   // Russell Crowe
	2227,  // Nicole Kidman
	190,   // Clint Eastwood
	3,     // Harrison Ford
	4173,  // Anthony Hopkins
	15735, // Helen Mirren

	// --- Classic Hollywood ---
	3636,  // Paul Newman
	3084,  // Marlon Brando
	4135,  // Robert Redford
	193,   // Gene Hackman
	1229,  // Jeff Bridges
	738,   // Sean Connery
	4690,  // Christopher Walken
	6949,  // John Malkovich
	1037,  // Harvey Keitel
	4517,  // Joe Pesci
	2228,  // Sean Penn
	15152, // James Earl Jones
	3460,  // Gene Wilder
	3087,  // Robert Duvall
	3392,  // Michael Douglas
	1269,  // Kevin Costner
	569,   // Ethan Hawke
	131,   // Jake Gyllenhaal
	138,   // Quentin Tarantino
	1243,  // Woody Allen

	// --- Comedy ---
	1327,  // Ian McKellen
	2387,  // Patrick Stewart
	67773, // Steve Martin
	2157,  // Robin Williams
	518,   // Danny DeVito
	9309,  // Richard Pryor
	7904,  // Billy Crystal
	206,   // Jim Carrey
	1532,  // Bill Murray
	7447,  // Alec Baldwin
	7180,  // John Candy
	10859, // Ryan Reynolds
	19292, // Adam Sandler
	12073, // Mike Myers
	12052, // Gwyneth Paltrow
	21007, // Jonah Hill
	19274, // Seth Rogen
	56323, // Tina Fey
	55536, // Melissa McCarthy

	// --- Action / Blockbuster ---
	16483, // Sylvester Stallone
	1100,  // Arnold Schwarzenegger
	2461,  // Mel Gibson
	776,   // Eddie Murphy
	18918, // Dwayne Johnson
	976,   // Jason Statham
	10814, // Wesley Snipes
	2963,  // Nicolas Cage
	2048,  // Gary Busey
	3896,  // Liam Neeson
	56731, // Jessica Alba
	15111, // Jean-Claude Van Damme
	64,    // Gary Oldman
	2176,  // Tommy Lee Jones
	17605, // Idris Elba
	1230,  // John Goodman

	// --- Bond Universe ---
	517,   // Pierce Brosnan
	8784,  // Daniel Craig
	5309,  // Judi Dench
	10669, // Timothy Dalton
	10222, // Roger Moore
	16940, // Jeremy Irons
	5469,  // Ralph Fiennes
	17838, // Rami Malek

	// --- Marvel / DC ---
	3223,   // Robert Downey Jr.
	16828,  // Chris Evans
	74568,  // Chris Hemsworth
	73457,  // Chris Pratt
	103,    // Mark Ruffalo
	17604,  // Jeremy Renner
	22226,  // Paul Rudd
	172069, // Chadwick Boseman
	1896,   // Don Cheadle
	8691,   // Zoe Saldaña
	17288,  // Michael Fassbender
	5530,   // James McAvoy
	10205,  // Sigourney Weaver
	30614,  // Ryan Gosling
	60073,  // Brie Larson

	// --- Prestige Drama ---
	504,   // Tim Robbins
	1979,  // Kevin Spacey
	884,   // Steve Buscemi
	3905,  // William H. Macy
	11701, // Angelina Jolie
	1461,  // George Clooney
	3895,  // Michael Caine
	5081,  // Emily Blunt
	1920,  // Winona Ryder
	4785,  // Jeff Goldblum
	1205,  // Richard Gere
	3036,  // John Cusack
	1233,  // Philip Seymour Hoffman
	3810,  // Javier Bardem
	3131,  // Antonio Banderas
	11856, // Daniel Day-Lewis
	1231,  // Julianne Moore
	350,   // Laura Linney
	4764,  // John C. Reilly
	4492,  // Philip Baker Hall
	18288, // Terrence Howard
	9777,  // Cuba Gooding Jr.

	// --- Awards Circuit ---
	54693,   // Emma Stone
	72129,   // Jennifer Lawrence
	19492,   // Viola Davis
	112,     // Cate Blanchett
	1813,    // Anne Hathaway
	9273,    // Amy Adams
	83002,   // Jessica Chastain
	1267329, // Lupita Nyong'o
	36592,   // Saoirse Ronan

	// --- International Stars ---
	3490,  // Adrien Brody
	1121,  // Benicio del Toro
	9642,  // Jude Law
	3061,  // Ewan McGregor
	5472,  // Colin Firth
	955,   // Penélope Cruz
	3136,  // Salma Hayek Pinault
	1922,  // Catherine Zeta-Jones
	2296,  // Clive Owen
	6162,  // Paul Bettany
	76788, // Dev Patel
	116,   // Keira Knightley

	// --- Character Actors ---
	5049,   // John Hurt
	227,    // William Hurt
	335,    // Michael Shannon
	932967, // Mahershala Ali
	5294,   // Chiwetel Ejiofor

	// --- TV Crossover ---
	17419, // Bryan Cranston
	4691,  // James Gandolfini
	65717, // Jon Hamm
	25072, // Oscar Isaac
	10297, // Matthew McConaughey

	// --- Women in Film ---
	18277, // Sandra Bullock
	368,   // Reese Witherspoon
	4587,  // Halle Berry
	204,   // Kate Winslet
	1038,  // Jodie Foster
	3141,  // Marisa Tomei
	139,   // Uma Thurman
	448,   // Hilary Swank
	3092,  // Diane Keaton
	4431,  // Jessica Lange
	5606,  // Sissy Spacek

	// --- New Generation ---
	1190668, // Timothée Chalamet
	206919,  // Daniel Kaluuya
	1373737, // Florence Pugh
	505710,  // Zendaya
	1136406, // Tom Holland
	132157,  // Ezra Miller
	2034418, // Jacob Elordi
	86654,   // Austin Butler
	1290466, // Barry Keoghan
}
