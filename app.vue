<script setup lang="ts">
import Pianoroll from './components/pianoroll.vue';
const actx = ref<AudioContext>();

const pianoRoll = ref<InstanceType<typeof Pianoroll> | null>(null);
const midiNotes = Array.from({ length: 128 }, (_, i) => i); // 0-127 midi notes

let melody = [];

const NoteDurations = {
  Whole: 16,
  Half: 8,
  Quarter: 4,
  Eighth: 2,
//   EighthTriplet: 3/2,
  Sixteenth: 1,
};

// add to types later
interface SequenceStep {
  t: number;
  n: any;
  g: number;
  f: number;
}



const selectedScale = ref<keyof typeof scaleIntervals>('minor');


const activeOscillators = ref(0);

const sequenceAutomated = ref<SequenceStep[]>([]);	// default

const compositionSteps = ref(32*4);

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
const selectedNote = ref(humanReadableNotes.value[48].value);
const selectedBassNote = ref(humanReadableNotes.value[24].value);
const selectedHarmonyNote = ref(humanReadableNotes.value[12].value);


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
	[0.2, 0.2, 0.2, 0.1, 0.1, 0.1, 0.1], // octave
];

const rhyTransitionMatrix = {
  Whole: { Whole: 0.1, Half: 0.4, Quarter: 0.3, Eighth: 0.2, Sixteenth: 0.0 },
  Half: { Whole: 0.2, Half: 0.3, Quarter: 0.4, Eighth: 0.1, Sixteenth: 0.0 },
  Quarter: { Whole: 0.05, Half: 0.1, Quarter: 0.4, Eighth: 0.3, Sixteenth: 0.15 },
  Eighth: { Whole: 0.02, Half: 0.05, Quarter: 0.2, Eighth: 0.5, Sixteenth: 0.23 },
  Sixteenth: { Whole: 0.01, Half: 0.02, Quarter: 0.05, Eighth: 0.1, Sixteenth: 0.82 },
};
const generateMelody = () => {
	const scale = generateScale(selectedNote.value+12);
	const melody = generateMarkovMelody(scale, transitionMatrix, compositionSteps.value, "melody");
	return melody;
}

const generateBass = () => {
	const scale = generateScale(selectedNote.value-24);
	const bass = generateMarkovMelody(scale, transitionMatrix, compositionSteps.value, "bass");
	return bass;
}

const generateHarmony = () => {
	const scale = generateScale(selectedNote.value-0);
	const harmony = generateMarkovMelody(scale, transitionMatrix, compositionSteps.value, "harmony");
	return harmony;
}

const generateScale = (rootNote: number) => {
	const intervals = scaleIntervals[selectedScale.value];
	const scale = [rootNote];
	let currentNote = rootNote;
	intervals.forEach((interval) => {
		currentNote += interval;
		scale.push(currentNote);
	});
	return scale;
}



// TODO: stuff to move to their own files
const selectedOscType = ref("sine");

const adsrParams = reactive({
	attack: { time: 0.05, volume: 0.1 },
	decay: { time: 0.01, volume: 0.1 },
	sustain: { time: 0.01, volume: 0.1 },
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
const tempo = ref(40);
const markEndPos = ref(32);
const markStartPos = ref(0);

//End TODO


/**
 * Generates a melody using a Markov chain based on the given scale and transition matrix.
 *
 * @param {string | any[]} scale - The musical scale to generate the melody from. Can be an array of notes or a string.
 * @param {any[]} matrix - The transition matrix representing the probabilities of moving from one note to another.
 * @param {number} [steps=16] - The number of steps (notes) in the generated melody. Defaults to 16.
 * @returns {any[]} - The generated melody as an array of notes.
 */
function generateMarkovMelody(scale: any[], matrix: any[], steps: number, type: string) {
	if(type == "melody") {
		 
	} else if(type == "bass") {
		
	} else if(type == "harmony") {

	}

	let melody: SequenceStep[] = [];
	let currentIndex = Math.floor(Math.random() * scale.length);

	for (let i = 0; i < steps; i++) {

		if (currentIndex == 7) {
			// this just bumps or dec octave
			for (let x = 0; x < scale.length; x++) {
				let oct = Math.random() > .5 ? 12 : -12
				scale[x] += oct;
			}
		}
		
		melody.push(scale[currentIndex]);

		// Choose next note based on probabilities
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

function generateRhythm(startState: keyof typeof NoteDurations, stepLimit: number) {
  const rhythm = [];
  let currentState = startState;
  let totalSteps = 0;

  while (totalSteps < stepLimit) {
    const duration = NoteDurations[currentState];
    if (totalSteps + duration > stepLimit) break;

    rhythm.push(duration);
    totalSteps += duration;

    currentState = selectNextState(currentState);
  }

  return rhythm;
}

function selectNextState(currentState: keyof typeof NoteDurations) {
  const probabilities = rhyTransitionMatrix[currentState];
  const random = Math.random();
  let cumulative = 0;

  for (const [state, probability] of Object.entries(probabilities)) {
    cumulative += probability;
    if (random < cumulative) {
      return state;
    }
  }
}

const play = async (ev: any) => {
	let gain = new GainNode(actx.value);//actx.value.createGain();
	let osc = new OscillatorNode(actx.value);
	// osc = actx.value.createOscillator();
	osc.onended = () => { // cleanup hook at end of playback
		activeOscillators.value--;
		osc.disconnect();
		gain.disconnect();
	};
	
	osc.type = selectedOscType.value as OscillatorType;
	osc.connect(gain).connect(actx.value.destination); // connect to output
	let frequency = 440 * Math.pow(2, (ev.n - 69) / 12);
	osc.frequency.setValueAtTime(frequency, ev.t); // note pitch
	gain.gain.linearRampToValueAtTime(0, ev.t);// init at 0

	let baseDuration = ev.t + (ev.g-ev.t)
	//+ (ev.g / tempo.value / 4); // base note
	// ADSR

	gain.gain.linearRampToValueAtTime(adsrParams.attack.volume, baseDuration*1000 +
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

const handleRandom = (): void => {
	const melody = generateMelody();
	const bass = generateBass();
	const harmony = generateHarmony();
	const rhy1 = generateRhythm("Whole", compositionSteps.value);
	
	for (let i=0, j=0; j < melody.length && i < melody.length;j+=rhy1[i], i++) {
		sequenceAutomated.value[i] = { t: j, n: melody[i], g: rhy1[i%rhy1.length], f: 0 };
		sequenceAutomated.value[i + rhy1.length] = { t: j, n: bass[i], g: rhy1[i%rhy1.length], f: 0 };
		sequenceAutomated.value[i + rhy1.length*2] = { t: j, n: harmony[i], g: rhy1[i%rhy1.length], f: 0 };
	}
	console.log(melody);
	console.log(bass);
	console.log(harmony);
	console.log(sequenceAutomated.value);
	
}

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
			:sequence="sequenceAutomated" :tempo="tempo" :xRange="200" :yRange="200" :xOffset="0" :yOffset="2"
			:timeBase="timeBase" :markEndPos="compositionSteps" :markStartPos="0">
		</Pianoroll>

		<button label="Play" @click="handlePlay">PLAY</button>
		<button label="Stop" @click="handleStop">STOP</button>
		<button label="Random" @click="handleRandom">RANDOM GENERATE</button>
		<!-- ADSR - to be componentized -->
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
