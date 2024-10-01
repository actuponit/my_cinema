export const cinemaFormat = (val: string) => {
  if (val !== 'IMAX') return '_'+val;
  return val;
}

export const cinemaFormatReverse = (val: string) => {
  if (val !== 'IMAX') return val;
  return val.slice(1);
}