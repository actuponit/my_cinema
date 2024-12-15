export default function displayImage(url: string | null | undefined) {
  if(!url)
    return "https://placehold.co/400"
  return "https://my-cinema-1.onrender.com/"+url
}