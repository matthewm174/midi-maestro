export class OscillatorPool {
    private audioContext: AudioContext;
    public poolSize: number;
    private oscillators: Array<{
        oscillator: OscillatorNode;
        gainNode: GainNode;
        inUse: boolean;
    }>;
    private inUse: Set<{
        oscillator: OscillatorNode;
        gainNode: GainNode;
        inUse: boolean;
    }>;

    constructor(audioContext: AudioContext, poolSize: number = 10) {
        this.audioContext = audioContext;
        this.poolSize = poolSize;
        this.oscillators = [];
        this.inUse = new Set();

        // Initialize the pool
        for (let i = 0; i < poolSize; i++) {
            const oscillator = this.audioContext.createOscillator();
            oscillator.type = "sine"; // Default type
            oscillator.frequency.setValueAtTime(440, this.audioContext.currentTime); // Default frequency

            const gainNode = this.audioContext.createGain();
            gainNode.gain.setValueAtTime(0, this.audioContext.currentTime); // Start muted

            oscillator.connect(gainNode).connect(this.audioContext.destination);
            oscillator.start();

            this.oscillators.push({ oscillator, gainNode, inUse: false });
        }
    }

    /**
     * Acquires an available oscillator from the pool.
     * @returns The oscillator object.
     * @throws Error if no oscillators are available.
     */
    acquireOscillator(): { oscillator: OscillatorNode; gainNode: GainNode } {
        for (const oscObj of this.oscillators) {
            if (!oscObj.inUse) {
                oscObj.inUse = true;
                this.inUse.add(oscObj);
                return { oscillator: oscObj.oscillator, gainNode: oscObj.gainNode };
            }
        }
        throw new Error("No available oscillators in the pool");
    }

    /**
     * Releases an oscillator back to the pool.
     * @param oscObj - The oscillator object to release.
     */
    releaseOscillator(oscObj: { oscillator: OscillatorNode; gainNode: GainNode }): void {
        const target = this.oscillators.find(
            (o) => o.oscillator === oscObj.oscillator && o.gainNode === oscObj.gainNode
        );

        if (target) {
            target.gainNode.gain.setValueAtTime(0, this.audioContext.currentTime); // Silence the oscillator
            target.inUse = false;
            this.inUse.delete(target);
        }
    }

    /**
     * Updates all in-use oscillators with a callback function.
     * @param callback - The function to update oscillator properties.
     */
    updateOscillators(
        callback: (oscillator: OscillatorNode, gainNode: GainNode) => void
    ): void {
        this.inUse.forEach((oscObj) => {
            callback(oscObj.oscillator, oscObj.gainNode);
        });
    }
}
// const audioContext = new AudioContext();
// const oscillatorPool = new OscillatorPool(audioContext, 5);
// // Acquire an oscillator and play a note
// const osc1 = oscillatorPool.acquireOscillator();
// osc1.oscillator.frequency.setValueAtTime(440, audioContext.currentTime); // Set frequency to A4
// osc1.gainNode.gain.setValueAtTime(0.5, audioContext.currentTime); // Set volume

// setTimeout(() => {
//     oscillatorPool.releaseOscillator(osc1); // Stop the note after 1 second
// }, 1000);

// // Update all in-use oscillators
// oscillatorPool.updateOscillators((oscillator, gainNode) => {
//     oscillator.frequency.setValueAtTime(660, audioContext.currentTime); // Change frequency
// });