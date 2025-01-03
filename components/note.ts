/**
 * Types file for various conversions
 */
export interface Note {
    c?: number;
    t: number; // Time in ticks
    g: number; // Duration in ticks
    n?: number; // Note or other identifier
    f?: number; // is the note selected?
    v?: number; // Velocity
    t1?: number; // Start tick
    t2?: number; // Next event's tick or markend
    dt?: number; // Delta between ticks
    i?: number;  // Index of the next event, or -1 if none found
    on?: number; // Optional property 'on' - original note before manipulation
    ot?: number; // Optional property 'ot' - original tick before manipulation
    og?: number; // Optional property 'og' - original duration before manipulation
    ev?: Note; // Optional property 'ev'
}

export interface MoveSelectedNote {
    dt: number;
    dn: number;
}