<script setup lang="ts">
import Pianoroll from './components/pianoroll.vue';
const actx = ref<AudioContext>();

const pianoRoll = ref<InstanceType<typeof Pianoroll> | null>(null);
const midiNotes = Array.from({ length: 128 }, (_, i) => i); // 0-127 midi notes

const NoteDurations = {
	Whole: 16,
	Half: 8,
	// QuartDot: 6,
	Quarter: 4,
	Eighth: 2,
	// EighthTriplet: 3/2, leaving complicated rhythms for later
	Sixteenth: 1,
};

// add to types later
interface SequenceStep {
	t: number;
	n: number;
	g: number;
	f: number;
}

const totalScaleNotes = ref<number[]>([]);


const selectedScale = ref<keyof typeof scaleIntervals>('minor');


const activeOscillators = ref(0);

const sequenceAutomated = ref<SequenceStep[]>([]);	// default
const sequenceAutomatedPreLoop = ref<SequenceStep[]>([]);	// default

const compositionSteps = ref(32 * 4 * 4);

const humanReadableNotes = computed(() => {
	const noteNames = ["C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"];
	return midiNotes.map((noteNumber) => {
		const octave = Math.floor(noteNumber / 12) - 1;
		const noteName = noteNames[noteNumber % 12];
		return {
			value: noteNumber,
			label: `${noteName}${octave}`,
		};
	});
});
const selectedNote = ref(humanReadableNotes.value[36].value);
const selectedBassNote = ref(humanReadableNotes.value[36].value);
const selectedHarmonyNote = ref(humanReadableNotes.value[36].value);


const scaleIntervals: { [key: string]: number[] } = {
	major: [2, 2, 1, 2, 2, 2, 1],
	minor: [2, 1, 2, 2, 1, 2, 2],
	harmonicMinor: [2, 1, 2, 2, 1, 3, 1],
	melodicMinor: [2, 1, 2, 2, 2, 2, 1],
	pentatonicMajor: [2, 2, 3, 2, 3],
	pentatonicMinor: [3, 2, 2, 3, 2],
	blues: [3, 2, 1, 1, 3, 2],
	diminished: [2, 1, 2, 1, 2, 1, 2, 1],
	wholeTone: [2, 2, 2, 2, 2, 2],
	chromatic: [1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1],
};


const rockGenre = {
  fifth: 0.75,
  third: 0.2,
  both: 0.05,
};

let transitionMatrix = [
	// c minor 7 notes of transition for markov chain... 
	// i need to add the learning function since this is sort of magic number-y - 
	// this will come after i have midi upload and interpretation done
	[0.1, 0.3, 0.2, 0.1, 0.1, 0.1, 0.1], // C - when note is C, choose one based on prob
	[0.1, 0.1, 0.3, 0.2, 0.2, 0.1, 0.0], // D - when note is D, choose next
	[0.1, 0.2, 0.1, 0.3, 0.1, 0.1, 0.1], // Eb - etc..
	[0.1, 0.1, 0.2, 0.1, 0.3, 0.2, 0.0], // F
	[0.2, 0.1, 0.1, 0.3, 0.1, 0.2, 0.0], // G
	[0.1, 0.2, 0.1, 0.1, 0.3, 0.1, 0.1], // Ab
	[0.1, 0.1, 0.3, 0.2, 0.1, 0.1, 0.1], // Bb
];

const bassTransitionMatrix = {
	Whole: { Whole: 0.5, Half: 0.4, Quarter: 0.10, Eighth: 0.0,  Sixteenth: 0.0 },
	Half: { Whole: 0.2, Half: 0.3, Quarter: 0.50, Eighth: 0.0,  Sixteenth: 0.0 },
	Quarter: { Whole: 0.3, Half: 0.2, Quarter: 0.5, Eighth: 0.0,  Sixteenth: 0.0 },
	Eighth: { Whole: 0.0, Half: 0.0, Quarter: 1.0, Eighth: 0.0,  Sixteenth: 0.0 },
	Sixteenth: { Whole: 0.0, Half: 0.0, Quarter: 1.0, Eighth: 0.0, Sixteenth: 0.00 },
};

const rootTransitionMatrix = {
	Whole: { Whole: 0.2, Half: 0.4, Quarter: 0.15, Eighth: 0.25,  Sixteenth: 0.0 },
	Half: { Whole: 0.2, Half: 0.15, Quarter: 0.40, Eighth: 0.25,  Sixteenth: 0.0 },
	Quarter: { Whole: 0.3, Half: 0.2, Quarter: 0.25, Eighth: 0.25,  Sixteenth: 0.0 },
	Eighth: { Whole: 0.4, Half: 0.1, Quarter: 0.25, Eighth: 0.25,  Sixteenth: 0.0 },
	Sixteenth: { Whole: 0.2, Half: 0.15, Quarter: 0.40, Eighth: 0.25, Sixteenth: 0.00 },
};
const rhyTransitionMatrix = {
	Whole: { Eighth: 0.0, Sixteenth: 0.0 },
	Half: { Eighth: 0.0, Sixteenth: 0.0 },
	Quarter: { Eighth: 0.0, Sixteenth: 0.0 },
	Eighth: { Eighth: 0.75, Sixteenth: 0.25 },
	Sixteenth: { Eighth: 0.10, Sixteenth: 0.90 },
};


const generateMelody = (root: number) => {
	totalScaleNotes.value = generateScale(root);
	const melody = generateMarkovMelody(transitionMatrix, compositionSteps.value, root);
	return melody;
}


const generateScale = (rootNote: number) => {
  const intervals = scaleIntervals[selectedScale.value];
  const scale = [rootNote];
  let currentNoteUp = rootNote;
  let currentNoteDown = rootNote;
  
  // Generate notes upwards (higher pitches)
  while (currentNoteUp <= 127) {
    intervals.forEach((interval) => {
      currentNoteUp += interval;
      if (currentNoteUp <= 127) {
        scale.push(currentNoteUp);
      }
    });
  }

  // Generate notes downwards (lower pitches)
  while (currentNoteDown >= 0) {
    intervals.reverse().forEach((interval) => {
      currentNoteDown -= interval;
      if (currentNoteDown >= 0) {
        scale.unshift(currentNoteDown); // Prepend to maintain the order
      }
    });
  }

  return scale;
};

// TODO: stuff to move to their own files
const selectedOscType = ref("sine");

const adsrParams = reactive({
	attack: { time: 0.05, volume: 0.04 },
	decay: { time: 0.01, volume: 0.04 },
	sustain: { time: 0.01, volume: 0.04 },
	release: { time: 0.01, volume: 0.0 },
});

const oscSettings = reactive({
	types: ["sine", "square", "sawtooth", "triangle"],

});
const sliderRanges = reactive({
	attack: { min: 0, max: 1, step: 0.01 },
	decay: { min: 0, max: 1, step: 0.01 },
	sustain: { min: 0, max: 1, step: 0.01 },
	release: { min: 0, max: 1, step: 0.01 },
});

const sliderUnits = reactive({
	attack: "s",
	decay: "s",
	sustain: "s",
	release: "s",
});

const timeBase = ref(16);
const tempo = ref(75);

//End TODO


/**
 * Generates a melody using a Markov chain based on the given scale and transition matrix.
 *
 * @param {string | any[]} scale - The musical scale to generate the melody from. Can be an array of notes or a string.
 * @param {any[]} matrix - The transition matrix representing the probabilities of moving from one note to another.
 * @param {number} [steps=16] - The number of steps (notes) in the generated melody. Defaults to 16.
 * @returns {any[]} - The generated melody as an array of notes.
 */
function generateMarkovMelody( matrix: any[], steps: number, root: number) {

	let melody: SequenceStep[] = [];
	let currentIndex = Math.floor(Math.random() * totalScaleNotes.value.length) %7;

	for (let i = 0; i < steps; i++) {

		if (currentIndex == 7) {
			// this just bumps or dec octave
			// for (let x = 0; x < totalScaleNotes.value.length; x++) {
			// 	let oct = Math.random() > .5 ? 12 : -12
			// 	totalScaleNotes.value[x] += oct;
			// }
		}

		melody.push({
			n: totalScaleNotes.value[currentIndex + root],
			t: 0,
			g: 0,
			f: 0
		});

		const probabilities = matrix[currentIndex];
		const rand = Math.random();
		let cumulative = 0;

		for (let j = 0; j < probabilities.length; j++) {
			cumulative += probabilities[j];
			if (rand <= cumulative) {
				currentIndex = j;
				break;
			}
		}
	}

	return melody;
}

const generateRhythmSection = (startState: keyof typeof NoteDurations) => {
  const root = [];
  let currentState = startState;

  const barLength = 4;
  const measureLength = Math.random() < 0.3 ? 32 : 32;

  let totalSteps = 0;

  while (totalSteps < compositionSteps.value) {
    let measure = []; // store the current measure's rhythm
    let measureSteps = 0;

    //generate one measure of rhythm
    while (measureSteps < measureLength) {
      const duration = NoteDurations[currentState];
      if (measureSteps + duration > measureLength) break;

      measure.push(duration);
      measureSteps += duration;
	  

      currentState = selectNextStateRhy(currentState);
    }

    // repeat the measure for the bar
    for (let i = 0; i < barLength; i++) {
      if (totalSteps + measureSteps > compositionSteps.value){
		totalSteps++;
		break;
	  } 

      root.push(...measure); // append the repeated measure
      totalSteps += measureSteps;
    }
  }

  return root;
}

function selectNextStateRhy(currentState: keyof typeof NoteDurations): keyof typeof NoteDurations {
	const probabilities = rhyTransitionMatrix[currentState];
	const random = Math.random();
	let cumulative = 0;

	for (const [state, probability] of Object.entries(probabilities)) {
		cumulative += probability;
		if (random < cumulative) {
			return state as keyof typeof NoteDurations;
		}
	}
	return currentState; // fallback to current state if no transition occurs
}


function generateBassSection(startState: keyof typeof NoteDurations) {
	const root = [];
	let currentState = startState;
	let totalSteps = 0;
	while (totalSteps < compositionSteps.value) {
		const duration = NoteDurations[currentState];
		if (totalSteps + duration > compositionSteps.value) break;

		root.push(duration);
		totalSteps += duration;

		currentState = selectNextStateBass(currentState);
	}
	return root;
}
function selectNextStateBass(currentState: keyof typeof NoteDurations): keyof typeof NoteDurations {
	const probabilities = bassTransitionMatrix[currentState];
	const random = Math.random();
	let cumulative = 0;

	for (const [state, probability] of Object.entries(probabilities)) {
		cumulative += probability;
		if (random < cumulative) {
			return state as keyof typeof NoteDurations;
		}
	}
	return currentState; // fallback to current state if no transition occurs
}


function generateRoot(startState: keyof typeof NoteDurations) {
	const root = [];
	let currentState = startState;
	let totalSteps = 0;
	while (totalSteps < compositionSteps.value) {
		const duration = NoteDurations[currentState];
		if (totalSteps + duration > compositionSteps.value) break;

		root.push(duration);
		totalSteps += duration;

		currentState = selectNextStateRoot(currentState);
	}
	return root;
}
function selectNextStateRoot(currentState: keyof typeof NoteDurations): keyof typeof NoteDurations {
	const probabilities = rootTransitionMatrix[currentState];
	const random = Math.random();
	let cumulative = 0;

	for (const [state, probability] of Object.entries(probabilities)) {
		cumulative += probability;
		if (random < cumulative) {
			return state as keyof typeof NoteDurations;
		}
	}
	return currentState; // fallback to current state if no transition occurs
}

const play = async (ev: any) => {
	let gain = new GainNode(actx.value!);//actx.value.createGain();
	let osc = new OscillatorNode(actx.value!);
	osc.onended = () => { // cleanup hook at end of playback
		activeOscillators.value--;
		osc.disconnect();
		gain.disconnect();
	};

	osc.type = selectedOscType.value as OscillatorType;
	osc.connect(gain).connect(actx.value!.destination); // connect to output
	let frequency = 440 * Math.pow(2, (ev.n - 69) / 12);
	osc.frequency.setValueAtTime(frequency, ev.t); // note pitch
	gain.gain.linearRampToValueAtTime(0, ev.t);// init at 0

	let baseDuration = ev.t + (ev.g - ev.t)
	// ADSR

	gain.gain.linearRampToValueAtTime(adsrParams.attack.volume, ev.t +
		adsrParams.attack.time); // Attack
	gain.gain.linearRampToValueAtTime(adsrParams.decay.volume, baseDuration +
		adsrParams.attack.time + adsrParams.decay.time); // Decay
	gain.gain.linearRampToValueAtTime(adsrParams.sustain.volume, baseDuration +
		adsrParams.attack.time + adsrParams.decay.time +
		adsrParams.sustain.time); // sustain
	gain.gain.linearRampToValueAtTime(adsrParams.release.volume, baseDuration
		+ adsrParams.attack.time + adsrParams.decay.time
		+ adsrParams.sustain.time + adsrParams.release.time); // Release
	osc.start(ev.t);
	activeOscillators.value++;
	//end slightly after to taper
	osc.stop(baseDuration + adsrParams.sustain.time + adsrParams.release.time + adsrParams.attack.time + adsrParams.decay.time + .1);
};


const handleStop = (): void => {
	if (pianoRoll.value && actx.value) {
		pianoRoll.value.stop();
		actx.value.suspend();
		console.log("stopped");

	}
}

function getHarmonicNote(baseNote: number, interval: string) {
	const idx = totalScaleNotes.value.indexOf(baseNote);
	// probably should just make this a map, but this is fine for now
	switch(interval) {
		case "thirteenth":
			return totalScaleNotes.value[idx + 13];
		case "twelveth":
			return totalScaleNotes.value[idx + 12];
		case "eleventh":
			return totalScaleNotes.value[idx + 11];
		case "tenth":
			return totalScaleNotes.value[idx + 10];
		case "ninth":
			return totalScaleNotes.value[idx + 9];
		case "octave":
			return totalScaleNotes.value[idx + 8];
		case "seventh":
			return totalScaleNotes.value[idx + 7];
		case "sixth":
			return totalScaleNotes.value[idx + 6];
		case "fifth":
			return totalScaleNotes.value[idx + 5];
		case "fourth":
			return totalScaleNotes.value[idx + 4];
		case "fifth":
			return totalScaleNotes.value[idx + 3];
		case "third":
			return totalScaleNotes.value[idx + 2];
		case "second":
			return totalScaleNotes.value[idx + 1];
		case "minthirteenth":
			return totalScaleNotes.value[idx - 13];
		case "mintwelveth":
			return totalScaleNotes.value[idx - 12];
		case "mineleventh":
			return totalScaleNotes.value[idx - 11];
		case "mintenth":
			return totalScaleNotes.value[idx - 10];
		case "minninth":
			return totalScaleNotes.value[idx - 9];
		case "minoctave":
			return totalScaleNotes.value[idx - 8];
		case "minseventh":
			return totalScaleNotes.value[idx - 7];
		case "minsixth":
			return totalScaleNotes.value[idx - 6];
		case "minfifth":
			return totalScaleNotes.value[idx - 5];
		case "minfourth":
			return totalScaleNotes.value[idx - 4];
		case "minfifth":
			return totalScaleNotes.value[idx - 3];
		case "minthird":
			return totalScaleNotes.value[idx - 2];
		case "minsecond":
			return totalScaleNotes.value[idx - 1];
		case "boctave":
			return totalScaleNotes.value[idx - 16];
		case "boctavefifth":
			return totalScaleNotes.value[idx - 16 + 5];
		case "boctavethird":
			return totalScaleNotes.value[idx - 16 + 3];
		default:
			return baseNote;
	}

}

const handleRandom = (): void => {
	const harmony = generateMelody(selectedNote.value);
	const rhy1 = generateRoot("Half");
	const rhy2 = generateRhythmSection("Sixteenth");
	const rhy3 = generateBassSection("Whole");


	//generateScale
	for (let i = 0, j = 0; j < harmony.length && i < harmony.length; j += rhy1[i], i++) {
		sequenceAutomatedPreLoop.value.push({ t: j, n: harmony[i].n, g: rhy1[i % rhy1.length], f: 0 });
		if(rhy1[i % rhy1.length] > 1){
			const randomValue = Math.random();

			if (randomValue < rockGenre.fifth) {
				sequenceAutomatedPreLoop.value.push({
				t: j,
				n: getHarmonicNote(harmony[i].n, "fifth"),
				g: rhy1[i % rhy1.length],
				f: 0,
			});
			} else if (randomValue < rockGenre.fifth + rockGenre.third) {
				sequenceAutomatedPreLoop.value.push({
				t: j,
				n: getHarmonicNote(harmony[i].n, "minoctave"),
				g: rhy1[i % rhy1.length],
				f: 0,
			});
			} else {
				sequenceAutomatedPreLoop.value.push({
				t: j,
				n: getHarmonicNote(harmony[i].n, "minthird"),
				g: rhy1[i % rhy1.length],
				f: 0,
			});
			sequenceAutomatedPreLoop.value.push({
				t: j,
				n: getHarmonicNote(harmony[i].n, "minseventh"),
				g: rhy1[i % rhy1.length],
				f: 0,
			});
			}
		}
	}
	//harmony
	for (let i = 0, j = 0; j < harmony.length && i < harmony.length; j += rhy2[i], i++) {
		sequenceAutomatedPreLoop.value.push({ t: j, n: getHarmonicNote(harmony[i].n, "minoctave"), g: rhy2[i % rhy2.length], f: 0 });
	}

	//bass
	for (let i = 0, j = 0; j < harmony.length && i < harmony.length; j += rhy3[i], i++) {
		sequenceAutomatedPreLoop.value.push({ t: j, n: getHarmonicNote(harmony[i].n, "boctave"), g: rhy3[i % rhy3.length], f: 0 });
	}
	//sectioning
	const midiLoop = saveChunks(sequenceAutomatedPreLoop.value, 32);
	for (let i = 0; i < midiLoop.length; i++) {
		sequenceAutomated.value.push({ t: midiLoop[i].t, n: midiLoop[i].n, g: midiLoop[i].g, f: midiLoop[i].f });
	}
	// AABCCBBAABBDDDD 
}


const order = ["A", "A", "B", "B", "C", "C", "B", "B", "A", "A", "D", "D", "D", "D"];


const saveChunks = (sequence: any[], barLength: number) => {

	const ranges = [
		{ start: 0, end: 32 },
		{ start: 32, end: 64 },
		{ start: 64, end: 96 }, 
		{ start: 96, end: 128 }, 
		{ start: 128, end: 192 },
	];

  const chunks: { [key: string]: any[] } = {};
  ranges.forEach((range, index) => {
    chunks[String.fromCharCode(65 + index)] = sequence.filter(
      event => event.t >= range.start && event.t < range.end
    );
  });

  let reorderedSequence: any[] = [];
  let currentTick = 0;

  order.forEach(chunkLabel => {
    const chunk = chunks[chunkLabel];
    chunk.forEach(event => {
      reorderedSequence.push({
        ...event,
        t: currentTick + (event.t - chunk[0].t),
        originalT: event.t
      });
    });
    currentTick += barLength;
  });

  return reorderedSequence;
};


const handlePlay = (): void => {
	if (pianoRoll.value && actx.value) {
		actx.value.resume();
		console.log("playing");
		pianoRoll.value.play(actx.value!, 0);
	}
}

/**
 * this method is called when a note is played from the pianoroll
 * @param ev 
 */
const Callback = async (ev: any) => {
	play(ev);
}

/**
 * init audio context and oscillator(s) - wip
 */
onMounted(() => {
	actx.value = new AudioContext();

	// actx.value = new (window.AudioContext)();
});

</script>



<template>
	<div>
		<Pianoroll ref="pianoRoll" :playCallback="Callback" :pianoHeight="700" :pianoWidth="300"
			:sequence="sequenceAutomated" :tempo="tempo" :xRange="512" :yRange="200" :xOffset="0" :yOffset="2"
			:timeBase="timeBase" :markEndPos="compositionSteps" :markStartPos="0">
		</Pianoroll>

		<button label="Play" @click="handlePlay">PLAY</button>
		<button label="Stop" @click="handleStop">STOP</button>
		<button label="Random" @click="handleRandom">RANDOM GENERATE</button>
		<!-- ADSR and others to be componentized -->
		<div>
			<h1>ADSR Envelope</h1>
			<div class="slider-container" v-for="(value, param) in adsrParams" :key="param">
				<input :id="param" type="range" :min="sliderRanges[param].min" :max="sliderRanges[param].max"
					:step="sliderRanges[param].step" v-model.number="value.time" />
				<span>{{ param }} time: {{ value.time }}</span>
			</div>
		</div>
		<div>
			<h1>Osc Settings</h1>
			<div class="slider-container">
				<select v-model="selectedOscType">
					<option v-for="type in oscSettings.types" :key="type" :value="type">{{ type }}</option>
				</select>
			</div>

		</div>
		<h1>Generative Settings</h1>
		<div>
			<label for="rootNote">Select Bass Root Note:</label>
			<select id="rootNote" v-model="selectedBassNote">
				<option v-for="note in humanReadableNotes" :key="note.value" :value="note.value">
					{{ note.label }}
				</option>
			</select>
		</div>
		<div>
			<label for="rootNote">Select Harmony Root Note:</label>
			<select id="rootNote" v-model="selectedHarmonyNote">
				<option v-for="note in humanReadableNotes" :key="note.value" :value="note.value">
					{{ note.label }}
				</option>
			</select>
		</div>
		<div>
			<label for="rootNote">Select Melody Root Note:</label>
			<select id="rootNote" v-model="selectedNote">
				<option v-for="note in humanReadableNotes" :key="note.value" :value="note.value">
					{{ note.label }}
				</option>
			</select>
		</div>
		<div>
			<label for="scale">Select Scale:</label>
			<select id="scale" v-model="selectedScale">
				<option v-for="(intervals, scale) in scaleIntervals" :key="scale" :value="scale">
					{{ scale }}
				</option>
			</select>
		</div>
		<div>
			<label for="compositionSteps">Composition Steps:</label>
			<input id="compositionSteps" type="number" v-model="compositionSteps" />
		</div>
		<div>
			<label for="tempo">Tempo:</label>
			<input id="tempo" type="number" v-model="tempo" />
		</div>
		<div>
			<label for="timeBase">Time Base:</label>
			<input id="timeBase" type="number" v-model="timeBase" />
		</div>
		<div>
			activeOscillators
			{{ activeOscillators }}
		</div>
	</div>
</template>
