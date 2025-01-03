<script setup lang="ts">
import Pianoroll from './components/pianoroll.vue';
const actx = ref<AudioContext>();

const pianoRoll = ref<InstanceType<typeof Pianoroll> | null>(null);
const isMinor = ref(true);
const midiNotes = Array.from({ length: 128 }, (_, i) => i); // 0-127 midi notes

let melody = [];
const sequenceAutomated = ref([]);	// default


const humanReadableNotes = computed(() => {
	const noteNames = ["C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"];
	return midiNotes.map((noteNumber) => {
		const octave = Math.floor(noteNumber / 12) - 1;
		const noteName = noteNames[noteNumber % 12];
		return {
			value: noteNumber, // MIDI note number
			label: `${noteName}${octave}`, // Human-readable note (e.g., C4)
		};
	});
});
const selectedNote = ref(humanReadableNotes.value[60].value);

const selectedScale = ref<keyof typeof scaleIntervals>('minor');

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

const generateMelody = () => {
	const scale = generateScale(selectedNote.value, isMinor.value);
	melody = generateMarkovMelody(scale, transitionMatrix);
	return melody;
}

const generateScale = (rootNote: number, isMinor: boolean) => {
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

//End TODO


/**
 * Generates a melody using a Markov chain based on the given scale and transition matrix.
 *
 * @param {string | any[]} scale - The musical scale to generate the melody from. Can be an array of notes or a string.
 * @param {any[]} matrix - The transition matrix representing the probabilities of moving from one note to another.
 * @param {number} [steps=16] - The number of steps (notes) in the generated melody. Defaults to 16.
 * @returns {any[]} - The generated melody as an array of notes.
 */
function generateMarkovMelody(scale: any[], matrix: any[], steps = 16) {
	let melody = [];
	let currentIndex = Math.floor(Math.random() * scale.length);

	for (let i = 0; i < steps; i++) {
		
		if(currentIndex==7){
			//bump or reduce octave randomly when moving onto 8th note
			
			for(let i = 0; i < scale.length; i++){
				let oct = Math.random()>.5?12:-12
				scale[i] += oct;
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


const play = (ev: any) => {
	if (!actx.value) {
		actx.value = new AudioContext();
	}
	let gain = new GainNode(actx.value);//actx.value.createGain();
	let osc = new OscillatorNode(actx.value!);
	osc = actx.value.createOscillator();
	osc.onended = () => { // cleanup hook at end of playback
		osc.disconnect();
		gain.disconnect();
	};
	osc.type = selectedOscType.value as OscillatorType;
	osc.connect(gain).connect(actx.value.destination); // connect to output
	osc.detune.setValueAtTime((ev.n - 69) * 100, ev.t); // note pitch
	gain.gain.linearRampToValueAtTime(0, ev.t);// init at 0

	let baseDuration = ev.t + (ev.g / 16 / 100);
	// ADSR

	gain.gain.linearRampToValueAtTime(adsrParams.attack.volume, baseDuration +
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
	//end slightly after to taper
	osc.stop(ev.t + adsrParams.sustain.time + adsrParams.release.time + adsrParams.attack.time + adsrParams.decay.time + .1);
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
	console.log(melody);
	for (let i = 0; i < melody.length; i++) {
		sequenceAutomated.value[i] = { t: i, n: melody[i], g: 1, f: 0 };
	}
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
const Callback = (ev: any) => {

	playOsc1(ev);
}

/**
 * init audio context and oscillator(s) - wip
 */
onMounted(() => {

	actx.value = new (window.AudioContext)();
});

function playOsc1(ev: any) {
	play(ev);
}
</script>



<template>
	<div>
		<Pianoroll ref="pianoRoll" :playCallback="Callback" :pianoHeight="700" :pianoWidth="300"
			:sequence="sequenceAutomated">
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
			<label for="rootNote">Select Root Note:</label>
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

	</div>
</template>
