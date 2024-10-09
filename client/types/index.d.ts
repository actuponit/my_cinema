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
    movie?: string;
    id: number;
    date: string;
    time: string;
    hall: string;
    format: string;
    price: number;
}

// export type Movie = {
//     id: number;
//     title: string;
//     thumbnail: string;
//     rating: number;
//     duration: string;
//     releaseDate: Date;
//     genre: string;
//     totalreviews: number;
// }

export type MovieHome = Movie & {
    price: number;
    scheduleDate: Date;
}

export type MovieDetail = Movie & {
    description: string;
    isBookMarked?: boolean;
    images?: string[];
    castAndCrew?: Person[];
    reviews?: Review[];
    schedules?: Schedule[];
}

export type ReviewForm = {
    rating: number;
    comment: string;
}

export type Ticket = {
    movieTitle: string;
    hall: string;
    thumbnail: string;
    price: number;
    amount: number;
    id: string;
    boughtAt: Date;
    movieTime: Date;
    isUpcoming: boolean;
    format: string;
}
export type MovieRole = {
    id: number;
    title: string;
    year: number;
}

export type PersonDetail = {
    name: string;
    photoUrl: string;
    role: string;
    bio: string;
    movies: MovieRole[];
}