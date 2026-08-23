export type DuplicateRisk = "high" | "medium" | "low" | "none";

export interface DuplicateCandidate {
    member_id: number;
    name: string;
    phone: string;
    congregation: string;
    score: number;
    risk: DuplicateRisk;
    matched_fields: string[];
}

export interface DuplicateCheckResult {
    has_possible_duplicates: boolean;
    highest_risk: DuplicateRisk;
    candidates: DuplicateCandidate[];
}

export const duplicateRiskLabels: Record<DuplicateRisk, string> = {
    high: "Alta probabilidade",
    medium: "Possível correspondência",
    low: "Baixa probabilidade",
    none: "Sem correspondência",
};

export const duplicateFieldLabels: Record<string, string> = {
    name: "Nome",
    phone: "Telefone",
    address: "Endereço",
    address_number: "Número",
    neighborhood: "Bairro",
    city: "Cidade",
    congregation: "Congregação",
};
