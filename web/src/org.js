async function getOrgUrl () {
  const response = await window.fetch('/organization', {
    credentials: 'same-origin'
  })
  if (!response.ok) {
    return
  }
  const json = await response.json()
  if (!json) {
    return
  }
  return json.url
}

export { getOrgUrl }
