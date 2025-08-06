// Simple notification utility for the MongoServerManager component
// This can be replaced with a more sophisticated notification system if available

export function showNotification(message: string, type: 'success' | 'error' | 'info' = 'info') {
  // Simple implementation using alert
  // You can replace this with your preferred notification library
  if (type === 'error') {
    console.error(message);
  } else {
    console.log(message);
  }
  
  // For now, we'll use alert for simplicity
  // In a real application, you'd use a toast library like vue-sonner or similar
  alert(message);
}