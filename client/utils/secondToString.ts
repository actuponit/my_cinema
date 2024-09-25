export type DurationReturn = {
    hours: number;
    minutes: number;
    seconds: number
}

function secondToObject(duration: number): DurationReturn {
    const hours = Math.floor(duration / 3600);
    const minutes = Math.floor((duration % 3600) / 60);
    const seconds = duration % 60;
    return {
        hours,
        minutes,
        seconds
    }
}

export default secondToObject;