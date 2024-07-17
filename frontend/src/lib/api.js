const API_HOST = import.meta.env.VITE_API_HOST || 'http://localhost';
const API_PORT = import.meta.env.VITE_API_PORT || '8080';
const API_BASE_URL = `${API_HOST}:${API_PORT}`;

export async function fetchChannels() {
  const response = await fetch(`${API_BASE_URL}/api/channels`);
  if (!response.ok) {
    throw new Error('Network response was not ok');
  }
  return response.json();
}

export async function fetchChannelItems(channelId) {
  const response = await fetch(`${API_BASE_URL}/api/channels/${channelId}/items`);
  if (!response.ok) {
    throw new Error('Network response was not ok');
  }
  return response.json();
}

export async function addChannel(channel) {
  const response = await fetch(`${API_BASE_URL}/api/channels`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(channel)
  });
  if (!response.ok) {
    throw new Error('Network response was not ok');
  }
  return response.json();
}

export async function deleteChannel(channelId) {
  const response = await fetch(`${API_BASE_URL}/api/channels/${channelId}`, {
    method: 'DELETE'
  });
  if (!response.ok) {
    throw new Error('Network response was not ok');
  }
}

export async function deleteItem(itemId) {
  const response = await fetch(`${API_BASE_URL}/api/items/${itemId}`, {
    method: 'DELETE'
  });
  if (!response.ok) {
    throw new Error('Network response was not ok');
  }
}
