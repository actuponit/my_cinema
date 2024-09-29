export default function displayImage(url: string | null | undefined) {
  if(!url)
    return "https://placehold.co/400"
  return "http://localhost:3000/"+url
}