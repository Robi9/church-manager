export function formatBrazilianPhone(value: string): string {
    if (!value || value.includes("*")) return value;

    let digits = value.replace(/\D/g, "");
    if (digits.startsWith("55") && (digits.length === 12 || digits.length === 13)) {
        digits = digits.slice(2);
    }
    digits = digits.slice(0, 11);

    if (!digits) return "";
    if (digits.length <= 2) return `(${digits}`;

    const areaCode = digits.slice(0, 2);
    const number = digits.slice(2);
    if (number.length <= 4) return `(${areaCode}) ${number}`;

    const prefix = number.slice(0, -4);
    const suffix = number.slice(-4);
    return `(${areaCode}) ${prefix}-${suffix}`;
}
