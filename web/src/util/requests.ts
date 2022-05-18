export function parsJsonLines(response: Response) {
  return response.text().then(function (text: string) {
    const lines = text.trim().split('\n');
    if (lines[0] === '') return [];
    return lines.map((line) => JSON.parse(line));
  });
}

export function checkStatus(response: any) {
  if (response.status >= 200 && response.status < 300) {
    return Promise.resolve(response);
  } else {
    return Promise.reject({
      status: response.status,
      message: response.statusText,
    });
  }
}
