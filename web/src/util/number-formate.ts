export function FormatNumberRegex(num: number) {
  return num.toString().replace(/(\d)(?=(\d{3})+(?!\d))/g, '$1,');
}

const nfObject = new Intl.NumberFormat('en-US');

export function FormatNumber(num: number | null) {
  if (num == null) return '-';
  return nfObject.format(num);
}
