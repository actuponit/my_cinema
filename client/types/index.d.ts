export type Review = {
    id: number;
    userName: string;
    userImage: string;
    rating: number;
    comment: string;
}

export type Person = {
    id: number;
    name: string;
    role: string;
    image: string;
}

export type Schedule = {
    id: number;
    date: string;
    time: string;
    hall: string;
    format: string;
    price: number;
}

export type Movie = {
    id: number;
    title: string;
    images: string[];
    rating: number;
    description: string;
    genre: string[];
    duration: string;
    releaseDate: string;
    castAndCrew: Person[];
    reviews: Review[];
    schedules: Schedule[];
}

export type ReviewForm = {
    rating: number;
    comment: string;
}
