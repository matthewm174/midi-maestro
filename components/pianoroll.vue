<script lang="ts" setup>
import { ref } from 'vue';
import { onKeyDown, onKeyStroke, onKeyUp, useMouse, useMouseInElement } from '@vueuse/core'
import MarkStart from './piano-roll/images/markstart.svg?component'
import MarkEnd from './piano-roll/images/markend.svg?component'
import Cursor from './piano-roll/images/playcursor.svg?component'
import type { Note, MoveSelectedNote } from './note.ts';
const mouseData = useMouse();


// const windowWidth = ref(window.innerWidth);
//refs
const longtaptimer = ref<NodeJS.Timeout | null>(null);
const longtapcount = ref(0);
// const playcallback = ref(null);
const swidth = ref<number>(0);
const sheight = ref<number>(500);
const keyboard = useTemplateRef('keyboard');
const semiflag = ref([6, 1, 0, 1, 0, 2, 1, 0, 1, 0, 1, 0]);
const menu = ref<HTMLElement>();
const body = ref<HTMLElement>();
const ctx = ref<CanvasRenderingContext2D | null>(null);
const rcMenu = ref({ x: 0, y: 0, width: 0, height: 0 });
const dragging = ref<{ i?: any, ev?: Note[], x?: any, y?: 
    any, offsy?: any, offsx?: any, o?: any, p?: any, p2?:
     any, t1?: any, t2?: any, n2?: any, m?: any, t?: any, 
     n?: any, n1?: any, g?: any, dt?: any }>(
    { i: null, ev: [] }
);
const lastx = ref(0);
const lasty = ref(0);
const ctrlDown = ref(false);
const press = ref(0);
const steph = ref(0);
const stepw = ref(0);
const markstartsrcStyle = ref('');
const markendsrcStyleLeft = ref('');
const cursorsrcStyle = ref('');

const cursorsrc = useTemplateRef('cursorsrc');

const markendsrc = useTemplateRef('markendsrc');
const markstartsrc = useTemplateRef('markstartsrc');//
const pianoRollCanvas = useTemplateRef('pianoRollCanvas');
const downpos = ref<{ x: number, y: number }>();
const downht = ref<any>(null);

const pianoRollCanvasStyle = ref({});
const rcTarget = ref();
const width = ref(500);
const height = ref(320);
const timebase = ref(16);
const editmode = ref('dragpoly');
const xrange = ref(16);
const yrange = ref(127);
const xoffset = ref(0);
const yoffset = ref(0); 
const grid = ref(1);
const snap = ref(1);
const xscroll = ref(0);
const yscroll = ref(0);
const gridnoteratio = ref(1);
const xruler = ref(24); // height of the x ruler
const yruler = ref(24); // width of the y ruler
const octadj = ref(-1);
const cursorPos = ref(0);
const defvelo = ref(100);
const collt = ref("#adg");
const coldk = ref("#adf");
const colgrid = ref("#afd");
const colnote = ref("#111");
const colnotesel = ref("#eee");
const colnoteborder = ref("#aaa");
const colnoteselborder = ref("#fff");
const colrulerbg = ref("#abc");
const colrulerfg = ref("#fff");
const colrulerborder = ref("#000");
const colselarea = ref("rgba(0,0,0,0.3)");
// const cursoroffset = ref(0);
const markstartoffset = ref(0);
const markendoffset = ref(24);
const kbwidth = ref(24);
const actxRef = ref<AudioContext | null>(null);
const sequence = ref<Note[]>([]);
const timer = ref<NodeJS.Timer | null>(null);
const timestack = ref<[number, number, number][]>([]); // [time, tick, tickToTimeRatio]
const time0 = ref(0);
const time1 = ref(0);
const tick0 = ref(0);
const tick1 = ref(0);
const tempo = ref(120); // Example default tempo
const markstartPos = ref(0);
const markendPos = ref(32);
// const preload = ref(0.0);
const index1 = ref(0);
const tick2time = ref(0);

// watch(() => width.value, () => layout());
// watch(windowWidth, (newValue, oldValue) => {
//       console.log(`Window width changed from ${oldValue} to ${newValue}`);
//       width.value = windowWidth.value;
//       // Add your logic here
//     });
const props = defineProps<{
    playCallback: (event: { t: number, g: number, n: number }) => void;
    pianoWidth: number;
    pianoHeight: number;
    sequence: Note[];
    yOffset: number;
    xOffset: number;
    yRange: number;
    markEndPos: number;
    markStartPos: number;
    xRange: number;
    tempo: number;
    timeBase: number;
}>();



watch(() => width.value, () => layout());
watch(() => height.value, () => layout());
watch(() => props.timeBase, () => {
    timebase.value = props.timeBase;
    layout();
});

watch(() => xoffset.value, () => {
    layout();
});
watch(() => yoffset.value, () => {
    layout();
});
watch(() => props.xRange, () => {
    xrange.value = props.xRange;
    layout();
});
watch(() => props.yRange, () => {
    yrange.value = props.yRange;
    layout();
});
watch(() => props.tempo, () => {
    tempo.value = props.tempo;
    updateTimer();
});
watch(() => props.sequence, () => {
    sequence.value = props.sequence;
    layout();
    redraw();
}, { deep: true });

watch(() => props.pianoHeight, () => {
    height.value = props.pianoHeight;
    layout();
    redraw();
}, { deep: true });

watch(() => props.tempo, () => {
    tempo.value = props.tempo;
    updateTimer();
});

watch(() => gridnoteratio.value, () => {
    updateTimer()
});
watch(() => cursorPos.value, () => {
    redrawMarker()
});
watch(() => markstartPos.value, () => {
    redrawMarker()
});
watch(() => markendPos.value, () => {
    redrawMarker()
});

onKeyDown('Shift', (e) => {
    ctrlDown.value = true;
})
onKeyDown('Delete', (e) => {
    delSelectedNote();
    redraw();
})

onKeyUp('Shift', (e) => {
    ctrlDown.value = false;
})


onMounted(() => {
    sequence.value = props.sequence;
    width.value = window.innerWidth;
    height.value = props.pianoHeight;
    tempo.value = props.tempo;
    timebase.value = props.timeBase;
    xrange.value = props.xRange;
    yrange.value = props.yRange;
    yoffset.value = props.yOffset;
    xoffset.value = props.xOffset;
    markendPos.value = props.markEndPos
    // height.value = window.innerHeight;
    if (pianoRollCanvas.value) {
        rcTarget.value = pianoRollCanvas.value.getBoundingClientRect();
        ctx.value = pianoRollCanvas.value.getContext('2d') as CanvasRenderingContext2D;
    }
    layout();
    redraw();
});








const delSelectedNote = () => {
    const l = sequence.value.length;
    for (let i = l - 1; i >= 0; --i) {
        const ev = sequence.value[i];
        if (ev.f)
            sequence.value.splice(i, 1);
    }
};


/**
 * 
 * @param pos hit test position - gets the mouse position and sets 'mode' to the hovered component
 */
const hitTest = (pos: { x: number; y: number; t?: any }): { t: number; n: number; i: number; m: string } => {
    const ht = { t: 0, n: 0, i: -1, m: " " };
    const l = sequence.value.length;

    ht.t = xoffset.value + ((pos.x - yruler.value - kbwidth.value) / swidth.value) * xrange.value;
    ht.n = yoffset.value - (pos.y - height.value) / steph.value;

    if (pos.y >= height.value || pos.x >= width.value) {
        return ht;
    }

    if (pos.y < xruler.value) {
        ht.m = "x"; // these blocks just set the hitbox for the x and y rulers
        return ht;
    }

    if (pos.x < yruler.value + kbwidth.value) {
        ht.m = "y";
        return ht;
    }

    for (let i = 0; i < l; ++i) {
        const ev = sequence.value[i];

        if ((ht.n | 0) === ev.n) {
            if (ev.f && Math.abs(ev.t - ht.t) * stepw.value < 8) {
                ht.m = "B"; // this block sets the hitbox for the beginning of a note
                ht.i = i;
                return ht;
            }
            if (ev.f && Math.abs(ev.t + ev.g - ht.t) * stepw.value < 8) {
                ht.m = "E"; // this block sets the hitbox for the end of a note
                ht.i = i;
                return ht;
            }
            if (ht.t >= ev.t && ht.t < ev.t + ev.g) {
                ht.i = i;
                // center of note
                ht.m = sequence.value[i].f ? "N" : "n";
                return ht;
            }
        }
    }

    ht.m = "s"; // none of the above, so it must be a space
    return ht;
};

const clearSel = (): void => {
    const l = sequence.value.length;
    for (let i = 0; i < l; ++i) {
        sequence.value[i].f = 0;
    }
};

const editDragDown = (pos: { x: number; y: number }): void => {
    const ht = hitTest(pos);
    let ev;
    if (ht.m === "N") {
        ev = sequence.value[ht.i];
        dragging.value = { o: "D", m: "N", i: ht.i, t: ht.t, n: ev.n, dt: ht.t - ev.t };
        for (let i = 0, l = sequence.value.length; i < l; ++i) {
            ev = sequence.value[i];
            if (ev.f) {
                ev.on = ev.n;
                ev.ot = ev.t;
                ev.og = ev.g;
            }
        }
        redraw();
    } else if (ht.m === "n") {
        
        ev = sequence.value[ht.i];
        clearSel();
        ev.f = 1;
        redraw();
    } else if (ht.m === "E") {
        ev = sequence.value[ht.i];
        dragging.value = { o: "D", m: "E", i: ht.i, t: ev.t, g: ev.g, ev: selectedNotes().map(sel => sel.ev) };
    } else if (ht.m === "B") {
        ev = sequence.value[ht.i];
        dragging.value = { o: "D", m: "B", i: ht.i, t: ev.t, g: ev.g, ev: selectedNotes().map(sel => sel.ev) };
    } else if (ht.m === "s" && ht.t >= 0) {
        clearSel();
        const t = ((ht.t / snap.value) | 0) * snap.value;
        sequence.value.push({
            t: t, n: ht.n | 0, g: 1, f: 1
        });
        dragging.value = {
            o: "D", m: "E",
            i: sequence.value.length - 1, 
            t: t, g: 1, 
            ev: [{
                t: t, g: 1, ev: sequence.value[sequence.value.length - 1]
            }]
        };

        redraw();
    }
};

const selectedNotes = (): Array<{ i: number; ev: any; t: number; g: number }> => {
    const obj: Array<{ i: number; ev: any; t: number; g: number }> = [];
    for (let i = sequence.value.length - 1; i >= 0; --i) {
        const ev = sequence.value[i];
        if (ev.f) {
            obj.push({ i: i, ev: ev, t: ev.t, g: ev.g });
        }
    }
    return obj;
};

const editDragMove = (pos: { x: number; y: number }): void => {
    const ht = hitTest(pos); // Type of ht should match the return type of hitTest
    let ev: Note | undefined; // Assumed type of a note from `this.sequence`
    let t: number;

    if (dragging.value.o === "D") {
        switch (dragging.value.m) {
            case "E": // End of note
                if (dragging.value.ev) {
                    const dt = ((Math.max(0, ht.t) / snap.value + 0.9) | 0) * snap.value - dragging.value.t - dragging.value.g;
                    const list = dragging.value.ev

                    for (let i = list.length - 1; i >= 0; --i) {
                        const ev = list[i];
                        ev.g = list[i].g + dt;
                        if (ev.g <= 0) {
                            ev.g = 1;
                        }
                    }
                }
                redraw();
                break;

            case "B": // Beginning of note
                if (dragging.value.ev) {
                    const dt = ((Math.max(0, ht.t) / snap.value + 0.9) | 0) * snap.value - dragging.value.t;
                    const list = dragging.value.ev

                    for (let i = list.length - 1; i >= 0; --i) {
                        const ev = list[i];
                        ev.t = list[i].t + dt;
                        ev.g = list[i].g - dt;
                        if (ev.g <= 0) ev.g = 1;
                        // if (editmove.value === "dragmono") {
                        //     delAreaNote(ev.t, ev.g);
                        // }
                    }
                }
                redraw();
                break;

            default:
                ev = sequence.value[dragging.value.i];
                t = ((Math.max(0, ht.t) / snap.value + 0.5) | 0) * snap.value;
                ev.g = ev.t + ev.g - t;
                ev.t = t;

                if (ev.g < 0) {
                    ev.t += ev.g;
                    ev.g = -ev.g;
                    dragging.value.m = "E";
                } else if (ev.g === 0) {
                    ev.t = t - 1;
                    ev.g = 1;
                }
                redraw();
                break;

            case "N":
                ev = sequence.value[dragging.value.i];
                if (ht.t && ht.n) {
                    moveSelectedNote({ dt: ht.t - dragging.value.t, dn: ht.n - dragging.value.n });
                }
                redraw();
                break;
        }
    }
};




/**
 * Moves the selected note by a specified time and pitch delta.
 *
 * @param {Object} param - The parameter object.
 * @param {number} param.dt - The delta time to move the note.
 * @param {number} param.dn - The delta pitch to move the note.
 * @returns {void}
 */
const moveSelectedNote = ({ dt, dn }: MoveSelectedNote): void => {
    const l = sequence.value.length;
    /**
     * Adjusts the timing and note values of events in the sequence.
     * 
     * The first loop iterates through the sequence to find the minimum offset time (dt) 
     * required to ensure all events have non-negative offset times.
     * 
     * The second loop adjusts each event's time (t) and note (n) based on the calculated dt 
     * and the snap value.
     * 
     * @param {Array} sequence.value - The array of event objects to be processed.
     * @param {number} sequence.value[].f - A flag indicating whether the event should be processed.
     * @param {number} sequence.value[].ot - The original time offset of the event.
     * @param {number} sequence.value[].on - The original note value of the event.
     * @param {number} snap.value - The value used to snap event times to a grid.
     * @param {number} dt - The minimum offset time to ensure non-negative event times.
     * @param {number} dn - The note value adjustment to be applied to each event.
     */
    for (let i = 0; i < l; ++i) {
        const ev = sequence.value[i];
        if (ev.f && ev.ot !== undefined && ev.ot + dt < 0) {
            dt = -ev.ot;
        }

    }
    for (let i = 0; i < l; ++i) {
        const ev = sequence.value[i];
        if (ev.f) {
            ev.t = (((ev.ot + dt) / snap.value + 0.5) | 0) * snap.value;
            ev.n = ev.on + dn;
        }
    }
};

const longtapcountup = (): void => {
    if (++longtapcount.value >= 18) {
        if (longtaptimer !== null) {
            if (longtaptimer.value !== null) {
                clearInterval(longtaptimer.value);
            }
        }
        switch (downht.value.m) {
            case "N":
            case "B":
            case "E":
                // delNote(downht.value.i);
                // popMenu(downpos.value);
                // dragging.value = { o: "m" };
                break;
        }
    }
};

const getPos = (e: MouseEvent | Touch): { t: EventTarget | null; x: number; y: number } => {
    let t: EventTarget | null = null;
    if (e) {
        t = e.target;
        lastx.value = e.clientX - rcTarget.value.left;
        lasty.value = e.clientY - rcTarget.value.top;
    }
    if (
        lastx.value >= rcMenu.value.x &&
        lastx.value < rcMenu.value.x + rcMenu.value.width &&
        lasty.value >= rcMenu.value.y &&
        lasty.value < rcMenu.value.y + rcMenu.value.height
    ) {
        t = menu.value ?? null;
    }
    return { t: t, x: lastx.value, y: lasty.value };
};

const pointerUp = (ev: MouseEvent | TouchEvent) => {
    cancel(ev);
}

const cancel = (ev: MouseEvent | TouchEvent) => {
    let e;
    if ('touches' in ev) {
        e = null
    } else {
        e = ev;
    }
    if (longtaptimer.value) {
        clearInterval(longtaptimer.value);
    }
    const pos = getPos(e);
    if (dragging.value.o == "m") {
        // this.menu.style.display="none";
        // this.rcMenu={x:0,y:0,width:0,height:0};
        // if(pos.t==this.menu)
        //     this.delSelectedNote();
        // this.redraw();
        console.log('menu action.. not implemented')
    }
    if (dragging.value.o == "A") {
        selAreaNote(dragging.value.t1, dragging.value.t2, dragging.value.n1, dragging.value.n2);
        dragging.value = { o: null };
        redraw();
    }
    if (editmode.value == "dragmono") {
        for (let ii = sequence.value.length - 1; ii >= 0; --ii) {
            const ev = sequence.value[ii];
            if (ev && ev.f) {
                delAreaNote(ev.t, ev.g, ii);
            }
        }
    }

    dragging.value = { o: null };
};


const selAreaNote=(t1: number,t2: number,n1: number,n2: number)=>{
            let t, i=0, e=sequence.value[i];
            if(n1>n2)
                t=n1,n1=n2,n2=t;
            if(t1>t2)
                t=t1,t1=t2,t2=t;
            while(e){
                if(e.t>=t1 && e.t<t2 && e.n>=n1 && e.n <= n2)
                    e.f=1;
                else
                    e.f=0;
                e=sequence.value[++i];
            }
        };

const pointerDown = (ev: MouseEvent | TouchEvent) => {
    let e: MouseEvent | Touch;
    if ('touches' in ev) {
        e = ev.touches[0];
        downpos.value = getPos(e)
    }
    else { e = ev; downpos.value = { x: ev.clientX, y: ev.clientY }; }
    rcTarget.value = pianoRollCanvas.value?.getBoundingClientRect() || null;


    downht.value = hitTest(downpos.value);

    longtapcount.value = 0;
    longtaptimer.value = setInterval(longtapcountup, 100);

    if ('button' in e && (e.button == 2 || e.ctrlKey)) {
        switch (downht.value.m) {
            case "N":
            case "B":
            case "E":
                dragging.value = { o: "m" };
                break;
            default:
                if (editmode.value == "dragmono" || editmode.value == "dragpoly")
                    dragging.value = { o: "A", p: downpos.value, p2: downpos.value, t1: downht.value.t, n1: downht.value.n };
                break;
        }
        ev.preventDefault();
        ev.stopPropagation();
        pianoRollCanvas.value?.focus();
        return false;
    }
    if (!e.target) return;
    // check if markend/start/cursor is being dragged
    switch ((e.target as Node).parentNode?.id) {
        case markendsrc.value._.attrs.id:
            dragging.value = { o: "E", x: downpos.value.x, m: markendPos.value };
            ev.preventDefault();
            ev.stopPropagation();
            return false;
        case markstartsrc.value._.attrs.id:
            dragging.value = { o: "S", x: downpos.value.x, m: markstartPos.value };
            ev.preventDefault();
            ev.stopPropagation();
            return false;
        case cursorsrc.value._.attrs.id:
            dragging.value = { o: "P", x: downpos.value.x, m: cursorPos.value };
            ev.preventDefault();
            ev.stopPropagation();
            return false;
    }
    dragging.value = { o: null, x: downpos.value.x, y: downpos.value.y, offsx: xoffset.value, offsy: yoffset.value }; // Adjust offsx and offsy as needed
    pianoRollCanvas.value?.focus();
    switch (editmode.value) {
        case "gridpoly":
        case "gridmono":
            editGridDown(downpos.value);
            break;
        case "dragpoly": // default
        case "dragmono":
            editDragDown(downpos.value);
            break;
    }
    press.value = 1;
    if (ev.preventDefault) ev.preventDefault();
    if (ev.stopPropagation) ev.stopPropagation();
    return false;
};


const mousemove = (e: MouseEvent): void => {
    if (dragging.value.o == null) {
        rcTarget.value = pianoRollCanvas.value?.getBoundingClientRect() || null;
        const pos = getPos(e);
        const ht = hitTest(pos);
        switch (ht.m) {
            case "E":
                pianoRollCanvas.value!.style.cursor = "e-resize";
                break;
            case "B":
                pianoRollCanvas.value!.style.cursor = "w-resize";
                break;
            case "N":
                pianoRollCanvas.value!.style.cursor = "move";
                break;
            case "n":
                pianoRollCanvas.value!.style.cursor = "pointer";
                break;
            case "s":
                pianoRollCanvas.value!.style.cursor = "pointer";
                break;
        }
    }
};


const handleMouseMove = (e: MouseEvent) => {
    pointerMove(e);
    mousemove(e);
}

const pointerMove = (ev: MouseEvent): boolean => {
    let e: MouseEvent | Touch;
    if (pianoRollCanvas.value) {
        rcTarget.value = pianoRollCanvas.value.getBoundingClientRect();
    }

    // if ((ev as TouchEvent).touches) {
    //     e = (ev as TouchEvent).touches[0];
    // } else {
    e = ev as MouseEvent;
    // }

    if (longtaptimer.value) {
        clearInterval(longtaptimer.value);
    }

    const pos = getPos(e);

    // pos.x = ev.clientX

    const ht = hitTest(pos);
    switch (dragging.value.o) { //what object is being dragged?
        case null:
            if (xscroll.value) {
                xoffset.value = dragging.value.offsx + (dragging.value.x - pos.x) * (xrange.value / width.value);
            }
            if (yscroll.value) {
                yoffset.value = dragging.value.offsy + (pos.y - dragging.value.y) * (yrange.value / height.value);
            }
            break;
        // case "m":
        //     if (ht.m === "m") {
        //         this.menu.style.background = "#ff6";
        //     } else {
        //         this.menu.style.background = "#eef";
        //     }
        //     break;
        case "A":
            dragging.value.p2 = pos;
            dragging.value.t2 = ht.t;
            dragging.value.n2 = ht.n;
            redraw();
            break;
        case "E": {
            const p = Math.max(1, (dragging.value.m + (pos.x - dragging.value.x) / stepw.value + 0.5) | 0);
            if (markstartPos.value >= p) {
                markstartPos.value = p - 1;
            }
            markendPos.value = p;
            break;
        }
        case "S": {
            const p = Math.max(0, (dragging.value.m + (pos.x - dragging.value.x) / stepw.value + 0.5) | 0);
            if (markendPos.value <= p) {
                markendPos.value = p + 1;
            }
            markstartPos.value = p;
            break;
        }
        case "P":
            cursorPos.value = Math.max(0, (dragging.value.m + (pos.x - dragging.value.x) / stepw.value + 0.5) | 0);
            break;
    }

    switch (editmode.value) {
        case "gridpoly":
        case "gridmono":
            editGridMove(pos);
            break;
        case "dragpoly":
        case "dragmono":
            editDragMove(pos);
            break;
    }

    ev.stopPropagation();
    return false;
};


const editGridMove = (pos: { x: number; y: number }): void => {
    const ht = hitTest(pos);

    if (dragging.value.o === "G") {
        switch (dragging.value.m) {
            case "1": {
                const px = Math.floor(ht.t);
                if (ht.m === "s") {
                    if (editmode.value === "gridmono") {
                        delAreaNote(px, 1, ht.i);
                    }
                    addNote(px, ht.n | 0, 1, defvelo.value, 0);
                }
                break;
            }
            case "0":
                if (ht.m === "n") {
                    delNote(ht.i);
                }
                break;
        }
    }
};

const delNote = (idx: number): void => {
    sequence.value.splice(idx, 1);
    redraw();
};

const editGridDown = (pos: { x: number; y: number }): void => {
    const ht = hitTest(pos);
    if (ht.m === "n") {
        delNote(ht.i);
        dragging.value = { o: "G", m: "0" };
    } else if (ht.m === "s" && ht.t >= 0) {
        const pt = Math.floor(ht.t);
        if (editmode.value === "gridmono") {
            delAreaNote(pt, 1, ht.i);
        }
        addNote(pt, ht.n | 0, 1, defvelo.value);
        dragging.value = { o: "G", m: "1" };
    }
};


const handleWheel = (ev: any) => {
    ev.stopPropagation();
    ev.preventDefault();
    if (ev.deltaY > 0) {
        //up
        if (ctrlDown.value) {
            xrange.value += 2
        } else {
            yrange.value += 2
        }
    }
    if (ev.deltaY < 0) {
        //down
        if (ctrlDown.value) {

            if (xrange.value > 4) {
                xrange.value -= 2
            }

        } else {
            if (yrange.value > 4) {
                yrange.value -= 2
            }
        }
    }
    redraw();
}


const redrawYRuler = (): void => {
    if (ctx.value && keyboard.value && yruler.value) {
        ctx.value.textAlign = "right";
        ctx.value.font = `${steph.value / 2}px 'sans-serif'`;
        ctx.value.fillStyle = colrulerbg.value;
        ctx.value.fillRect(0, xruler.value, yruler.value, sheight.value);

        ctx.value.fillStyle = colrulerborder.value;
        ctx.value.fillRect(0, xruler.value, 1, sheight.value); // Left border
        ctx.value.fillRect(yruler.value, xruler.value, 1, sheight.value); // Right border
        ctx.value.fillRect(0, height.value - 1, yruler.value, 1); // Bottom border

        ctx.value.fillStyle = colrulerfg.value;
        for (let y = 0; y < 128; y += 12) {
            const ys = height.value - steph.value * (y - yoffset.value);
            ctx.value.fillRect(0, ys | 0, yruler.value, -1); // Tick mark
            ctx.value.fillText(`C${((y / 12) | 0) + octadj.value}`, yruler.value - 4, ys - 4);
        }

    }
}
const redrawXRuler = (): void => {
    if (ctx.value && xruler.value) {
        ctx.value.textAlign = "left";
        ctx.value.font = `${xruler.value / 2}px 'sans-serif'`;
        ctx.value.fillStyle = colrulerbg.value;
        ctx.value.fillRect(0, 0, width.value, xruler.value);

        ctx.value.fillStyle = colrulerborder.value;
        ctx.value.fillRect(0, 0, width.value, 1); // Top border
        ctx.value.fillRect(0, 0, 1, xruler.value); // Left border
        ctx.value.fillRect(0, xruler.value - 1, width.value, 1); // Bottom border
        ctx.value.fillRect(width.value - 1, 0, 1, xruler.value); // Right border

        ctx.value.fillStyle = colrulerfg.value;
        for (let t = 0; ; t += timebase.value) {
            const x = (t - xoffset.value) * stepw.value + yruler.value + kbwidth.value;
            ctx.value.fillRect(x | 0, 0, 1, xruler.value); // Vertical grid line
            ctx.value.fillText((t / timebase.value + 1).toString(), x + 4, xruler.value - 8); // Time label
            if (x >= width.value) break;
        }
    }
}
const redrawAreaSel = (): void => {
    if (dragging.value && dragging.value.o == "A" && ctx.value) {
        ctx.value.fillStyle = colselarea.value;
        ctx.value.fillRect(dragging.value.p.x, dragging.value.p.y, dragging.value.p2.x - dragging.value.p.x, dragging.value.p2.y - dragging.value.p.y);
    }
}
const redrawMarker = (): void => {
    if (markstartsrc.value && cursorsrc.value && markendsrc.value) {
        const cur = (cursorPos.value + xoffset.value) * stepw.value + yruler.value + kbwidth.value;
        cursorsrcStyle.value = `${cur}px`;

        const start = (markstartPos.value - xoffset.value) * stepw.value + yruler.value + kbwidth.value;
        markstartsrcStyle.value = `${start + markstartoffset.value}px`;

        const end = (markendPos.value - xoffset.value) * stepw.value + yruler.value + kbwidth.value;
        markendsrcStyleLeft.value = `${end - markendoffset.value}px`; // this compensates for the size of the image
    }

}

const redrawGrid = (): void => {
    if (ctx.value && swidth.value) {
        for (let y = 0; y < 128; ++y) {
            if (semiflag.value[y % 12] & 1)
                ctx.value.fillStyle = coldk.value;
            else
                ctx.value.fillStyle = collt.value;

            const ys = height.value - (y - yoffset.value) * steph.value;
            ctx.value.fillRect(yruler.value + kbwidth.value, ys | 0, swidth.value, -steph.value);
            ctx.value.fillStyle = colgrid.value;// #5555
            ctx.value.fillRect(yruler.value + kbwidth.value, ys | 0, swidth.value, 1);
        }

        for (let t = 0; ; t += grid.value) {
            const x = stepw.value * (t - xoffset.value) + yruler.value + kbwidth.value;
            ctx.value.fillRect(x | 0, xruler.value, 1, sheight.value);
            if (x >= width.value) break;
        }
    }

}

const redrawKeyboard = (): void => {
    if (yruler.value && ctx.value) {
        ctx.value.textAlign = "right";
        ctx.value.font = (steph.value / 2) + "px 'sans-serif'";
        ctx.value.fillStyle = "#fff";
        ctx.value.fillRect(yruler.value, xruler.value, yruler.value, sheight.value);
        ctx.value.fillStyle = "#000";
        for (let y = 0; y < 128; ++y) {
            const ys = height.value - steph.value * (y - yoffset.value);
            const ysemi = y % 12;
            const fsemi = semiflag.value[ysemi];
            if (fsemi & 1) {
                ctx.value.fillRect(yruler.value, ys, yruler.value / 2, -steph.value);
                ctx.value.fillRect(yruler.value, (ys - steph.value / 2) | 0, yruler.value, -1);
            }
            if (fsemi & 2)
                ctx.value.fillRect(yruler.value, ys | 0, yruler.value, -1);
            if (fsemi & 4)
                ctx.value.fillText("C" + (((y / 12) | 0) + octadj.value), yruler.value - 4, ys - 4);
        }
        ctx.value.fillRect(yruler.value, xruler.value, 1, sheight.value);
    }
};


const addNote = (tick: any, note: any, duration: any, vel: any, f?: any) => {
    if (tick >= 0 && note >= 0 && note < 128) {
        const ev: Note = { t: tick, c: 0x90, n: note, g: duration, v: vel, f: f };
        sequence.value.push(ev);
        sortSequence();
        redraw();
        return ev;
    }
    return null;
};

const sortSequence = () => {
    sequence.value.sort((x, y) => { return x.t - y.t; });
};

const redraw = (): void => {
    let x,w,y,x2,y2;
    if (ctx.value) {
        ctx.value.clearRect(0, 0, width.value, height.value);
        stepw.value = swidth.value / xrange.value;
        steph.value = sheight.value / yrange.value;

        redrawGrid();
        const notelen = 1;

        sequence.value.forEach((ev) => {
            w = ev.g * stepw.value;
            x = (ev.t - xoffset.value) * stepw.value + yruler.value + kbwidth.value;
            x2 = (x + w) | 0;
            y = height.value - (ev.n - yoffset.value) * steph.value;
            y2 = (y - steph.value) | 0;
            if (ctx.value) {
                ctx.value.fillStyle = ev.f ? colnotesel.value : colnote.value;
                ctx.value.fillRect(x | 0, y | 0, x2 - x * notelen, y2 - y);
                ctx.value.fillStyle = ev.f ? colnoteselborder.value : colnoteborder.value;
                ctx.value.strokeRect(x, y, x2 - x, y2 - y);
            }
        });
        redrawKeyboard();
        redrawYRuler();
        redrawXRuler();
        redrawAreaSel();
        redrawMarker();

    }
}




const layout = () => {
    if (!kbwidth.value || !pianoRollCanvas.value || !body.value)
        return;

    if (width.value) {
        pianoRollCanvas.value.width = width.value;
        body.value.style.width = pianoRollCanvas.value.style.width = width.value + "px";
    }
    if (height.value) {
        pianoRollCanvas.value.height = height.value;
        body.value.style.height = pianoRollCanvas.value.style.height = height.value + "px";
    }
    swidth.value = pianoRollCanvas.value.width - yruler.value;
    swidth.value -= kbwidth.value;
    sheight.value = pianoRollCanvas.value.width - xruler.value;
    redraw();
};

const locate = (tick: number): void => {
    cursorPos.value = tick;
};

const updateTimer = (): void => {
    tick2time.value = 4 * 60 / tempo.value / timebase.value;
};


const stop = () => {
    if (timer.value) {
        clearInterval(1);
    }
    timer.value = null;
};

/**
 * This function handles the timing and playback of events in a piano roll.
 * It updates the cursor position and manages the event stack based on the current audio context time.
 * 
 * The function performs the following steps:
 * 1. Checks if the audio context reference is available.
 * 2. Updates the cursor position based on the current time and the first event in the timestack.
 * 3. Redraws the marker.
 * 4. Iterates through the events in the sequence and updates the timestack and cursor position accordingly.
 * 5. Calls the playCallback function with the current event details.
 * 
 * The function uses the following reactive references and properties:
 * - actxRef: Reference to the audio context.
 * - timestack: Stack of timing events.
 * - cursorPos: Current position of the cursor.
 * - time0, time1: Timing references for the current and next events.
 * - tick0, tick1: Tick references for the current and next events.
 * - sequence: Array of events to be played.
 * - index1: Index of the current event in the sequence.
 * - markstartPos, markendPos: Start and end positions of the marker.
 * - tick2time: Conversion factor from ticks to time.
 * - editmode: Mode of the editor (e.g., gridmono, gridpoly).
 * - gridnoteratio: Ratio for grid note length.
 * - props.playCallback: Callback function to play the notes.
 * 
 * The function also uses the following helper functions:
 * - redrawMarker: Function to redraw the marker.
 * - findNextEv: Function to find the next event in the sequence.
 */

function interval() {
    if (!actxRef.value) return;
    const current = actxRef.value.currentTime;
    while (timestack.value.length > 1 && current >= timestack.value[1][0]) {
        timestack.value.shift();
    }
    cursorPos.value = timestack.value[0][1] + (current - timestack.value[0][0]) / timestack.value[0][2];
    redrawMarker();
    while (current >= time1.value) {
        time0.value = time1.value;
        tick0.value = tick1.value;
        let e = sequence.value[index1.value];
        if (!e || e.t >= markendPos.value) {
            timestack.value.push([time1.value, markstartPos.value, tick2time.value]);
            const p = findNextEv(markstartPos.value);
            time1.value += p.dt * tick2time.value;
            index1.value = p.i;
        }
        else {
            tick1.value = e.t;
            timestack.value.push([time1.value, e.t, tick2time.value]);
            let gmax = Math.min(e.t + e.g, markendPos.value) - e.t;
            if (editmode.value == "gridmono" || editmode.value == "gridpoly") {
                gmax *= gridnoteratio.value;
            }
            //  { t: time1.value, g: time1.value + gmax * tick2time.value, n: e.n };
            //call back to whatever is playing the notes
            // console.log(sequence.value)
            sequence.value.filter(obj => obj.t === e.t).forEach(obj => 
            {
                const callbackev = { t: time1.value, g: time1.value + gmax * tick2time.value, n: obj.n };
                props.playCallback(callbackev);
                // obj.playCallback(callbackev);
            });
            // props.playCallback(callbackev);
            e = sequence.value[++index1.value];
            if (!e || e.t >= markendPos.value) {
                time1.value += (markendPos.value - tick1.value) * tick2time.value;
                const p = findNextEv(markstartPos.value);
                timestack.value.push([time1.value, markstartPos.value, tick2time.value]);
                time1.value += p.dt * tick2time.value;
                index1.value = p.i;
            }
            else
                time1.value += (e.t - tick1.value) * tick2time.value;
        }
    }
}

const play = (actx: AudioContext, tick?: number) => {
    if (tick !== undefined) {
        locate(tick);
    }

    if (timer.value !== null) return;

    actxRef.value = actx;
    timestack.value = [];
    time0.value = time1.value = actxRef.value.currentTime + 0.1;
    tick0.value = tick1.value = cursorPos.value;
    tick2time.value = 4 * 60 / tempo.value / timebase.value;

    const p = findNextEv(cursorPos.value);
    index1.value = p.i;
    timestack.value.push([0, cursorPos.value, 0]);
    timestack.value.push([time0.value, cursorPos.value, tick2time.value]);
    time1.value += p.dt * tick2time.value;

    if (p.i < 0) {
        timestack.value.push([time1.value, markstartPos.value, tick2time.value]);
    } else {
        timestack.value.push([time1.value, p.t1, tick2time.value]);
    }

    // Set the interval
    timer.value = setInterval(() => interval(), 25);
};

defineExpose({ play, stop });


const findNextEv = (tick: number): { i: any, t1: any, dt: any, t2: any } => {
    for (let i = 0; i < sequence.value.length; ++i) {
        const nextev = sequence.value[i];
        if (nextev.t >= markendPos.value) {
            return { t1: tick, t2: markendPos.value, dt: markendPos.value - tick, i: -1 };
        }
        if (nextev.t >= tick) {
            return { t1: tick, t2: nextev.t, dt: nextev.t - tick, i: i };
        }
    }
    return { t1: tick, t2: markendPos.value, dt: markendPos.value - tick, i: -1 };
};



const delAreaNote = (t: any, g: any, n?: any) => {
    const l = sequence.value.length;
    for (let i = l - 1; i >= 0; --i) {
        const ev = sequence.value[i];
        if (typeof (n) != "undefined" && n != i) {
            if (t <= ev.t && t + g >= ev.t + ev.g) {
                sequence.value.splice(i, 1);
            }
            else if (t <= ev.t && t + g > ev.t && t + g < ev.t + ev.g) {
                ev.g = ev.t + ev.g - (t + g);
                ev.t = t + g;
            }
            else if (t >= ev.t && t < ev.t + ev.g && t + g >= ev.t + ev.g) {
                ev.g = t - ev.t;
            }
            else if (t > ev.t && t + g < ev.t + ev.g) {
                addNote(t + g, ev.n, ev.t + ev.g - t - g, defvelo.value);
                ev.g = t - ev.t;
            }
        }
    }
}



</script>
<template>
    <div class="wac-body" id="wac-body" touch-action="none" ref="body" @mousedown="pointerDown" @mouseup="pointerUp"
        @mousemove="handleMouseMove">
        <canvas id="wac-pianoroll" ref="pianoRollCanvas" touch-action="none" tabindex="0" :style="pianoRollCanvasStyle"
            @wheel="handleWheel"></canvas>
        <div :style="{
            // top: `${xruler}px`,
            height: `${sheight}px`,
            left: `${yruler}px`,
            width: `${kbwidth}px`,
            backgroundPosition: `0px ${sheight + steph * yoffset}px`,
            backgroundSize: `100% ${steph * 12}px`
        }" id="wac-kb" ref="keyboard">

        </div>
        <MarkEnd id="wac-proll-mark-end" :style="{ left: markendsrcStyleLeft }" class="marker" ref="markendsrc" />
        <MarkStart id="wac-proll-mark-start" :style="{ left: markstartsrcStyle }" class="marker" ref="markstartsrc" />
        <Cursor id="wac-proll-cursor" :style="{ left: cursorsrcStyle }" class="marker" ref="cursorsrc" />
        <div id="wac-menu" ref="menu">Delete</div>
    </div>
</template>

<style>
.pianoroll {
    background: #ccc;
}

:host {
    user-select: none;
    display: inline-block;
    font-family: sans-serif;
    font-size: 11px;
    padding: 0;
    margin: 0;
}

#wac-body {
    position: relative;
    margin: 0;
    padding: 0;
    width: 100%;
    height: 100%;
    overflow: hidden;
}

#wac-pianoroll {
    cursor: pointer;
    margin: 0;
    padding: 0;
    width: 100%;
    height: 100%;
    background-size: 100% calc(100%*12/16);
    background-position: left bottom;
}

#wac-menu {
    display: none;
    position: absolute;
    top: 0px;
    left: 0px;
    background: #eef;
    color: #000;
    padding: 2px 10px;
    border: 1px solid #66f;
    border-radius: 4px;
    cursor: pointer;
}

.marker {
    height: 24px;
    position: absolute;
    left: 0px;
    top: 0px;
    cursor: ew-resize;
}

#wac-kb {
    position: absolute;
    left: 0px;
    top: 0px;
    width: 100px;
    height: 100%;
    background: repeat-y;
    background-size: 100% calc(100%*12/16);
    background-position: left bottom;
}
</style>