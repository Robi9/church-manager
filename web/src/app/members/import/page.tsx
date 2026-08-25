"use client";

import { useState } from "react";
import Link from "next/link";
import { ArrowLeft, Upload, Download, Loader2, AlertTriangle } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { useAuth } from "@/lib/auth";
import { formatBrazilianPhone } from "@/lib/phone";
import {
    duplicateFieldLabels,
    type DuplicateCandidate,
} from "@/lib/member-duplicates";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api";

interface ImportError {
    row: number;
    error: string;
    code?: string;
    data?: string[];
    candidates?: DuplicateCandidate[];
}

interface ImportResult {
    imported: number;
    failed: number;
    job_id?: string;
    errors?: ImportError[];
}

interface ReviewNotice {
    row: number;
    kind: "success" | "dismissed";
    message: string;
}

export default function ImportMembersPage() {
    const { token } = useAuth();

    const [file, setFile] = useState<File | null>(null);
    const [loading, setLoading] = useState(false);
    const [result, setResult] = useState<ImportResult | null>(null);
    const [error, setError] = useState("");
    const [confirmingRow, setConfirmingRow] = useState<number | null>(null);
    const [dismissingRow, setDismissingRow] = useState<number | null>(null);
    const [reviewError, setReviewError] = useState<ImportError | null>(null);
    const [reviewNotices, setReviewNotices] = useState<ReviewNotice[]>([]);

    async function handleImport() {
        setError("");
        if (!file) {
            setError("Selecione um arquivo CSV antes de importar.");
            return;
        }

        setLoading(true);

        try {
            const formData = new FormData();
            formData.append("file", file);

            const res = await fetch(`${API_URL}/members/import`, {
                method: "POST",
                headers: {
                    Authorization: `Bearer ${token}`,
                },
                body: formData,
            });

            const json = await res.json();

            if (!res.ok) {
                throw new Error(json.error || "Erro ao importar arquivo");
            }

            const importResult = json.data as ImportResult;
            setResult(importResult);
            setReviewNotices([]);
            setReviewError(
                importResult.errors?.find((item) => item.code === "possible_duplicate") ?? null,
            );
        } catch (err) {
            console.error(err);
            setError(err instanceof Error ? err.message : "Erro ao importar arquivo");
        } finally {
            setLoading(false);
        }
    }

    async function downloadErrors(jobId: string) {
        try {
            const res = await fetch(
                `${API_URL}/members/import/errors/${jobId}`,
                {
                    headers: {
                        Authorization: `Bearer ${token}`,
                    },
                }
            );

            if (!res.ok) {
                throw new Error("Erro ao baixar arquivo");
            }

            const blob = await res.blob();

            const url = window.URL.createObjectURL(blob);

            const a = document.createElement("a");
            a.href = url;
            a.download = "erros_importacao.csv";

            document.body.appendChild(a);
            a.click();
            a.remove();

            window.URL.revokeObjectURL(url);
        } catch (err) {
            console.error(err);
            alert("Não foi possível baixar o relatório");
        }
    }

    async function downloadTemplate() {
        const res = await fetch(
            `${API_URL}/members/import/template`,
            {
                headers: {
                    Authorization: `Bearer ${token}`,
                },
            }
        );

        const blob = await res.blob();

        const url = window.URL.createObjectURL(blob);

        const a = document.createElement("a");
        a.href = url;
        a.download = "modelo_membros.csv";
        a.click();

        window.URL.revokeObjectURL(url);
    }

    async function confirmDuplicate(row: number) {
        const jobID = result?.job_id;
        if (!jobID) return;

        setError("");
        setConfirmingRow(row);
        try {
            const res = await fetch(
                `${API_URL}/members/import/errors/${jobID}/confirm`,
                {
                    method: "POST",
                    headers: {
                        Authorization: `Bearer ${token}`,
                        "Content-Type": "application/json",
                    },
                    body: JSON.stringify({ row }),
                },
            );
            const json = await res.json();
            if (!res.ok) {
                throw new Error(json.error || "Não foi possível confirmar o cadastro");
            }
            removePendingRow(row, true);
            setReviewNotices((current) => [...current, {
                row,
                kind: "success",
                message: `Linha ${row}: membro cadastrado com sucesso.`,
            }]);
            setReviewError(null);
        } catch (err) {
            setError(err instanceof Error ? err.message : "Não foi possível confirmar o cadastro");
        } finally {
            setConfirmingRow(null);
        }
    }

    async function dismissDuplicate(row: number) {
        const jobID = result?.job_id;
        if (!jobID) return;

        setError("");
        setDismissingRow(row);
        try {
            const res = await fetch(
                `${API_URL}/members/import/errors/${jobID}/dismiss`,
                {
                    method: "POST",
                    headers: {
                        Authorization: `Bearer ${token}`,
                        "Content-Type": "application/json",
                    },
                    body: JSON.stringify({ row }),
                },
            );
            const json = await res.json();
            if (!res.ok) {
                throw new Error(json.error || "Não foi possível descartar a linha");
            }
            removePendingRow(row, false);
            setReviewNotices((current) => [...current, {
                row,
                kind: "dismissed",
                message: `Linha ${row}: cadastro cancelado e removido da revisão.`,
            }]);
            setReviewError(null);
        } catch (err) {
            setError(err instanceof Error ? err.message : "Não foi possível descartar a linha");
        } finally {
            setDismissingRow(null);
        }
    }

    function removePendingRow(row: number, imported: boolean) {
        setResult((current) => current ? {
            ...current,
            imported: current.imported + (imported ? 1 : 0),
            failed: Math.max(0, current.failed - 1),
            errors: current.errors?.filter((item) => item.row !== row),
        } : current);
    }

    const importErrors = result?.errors ?? [];

    return (
        <div className="space-y-6">
            <div className="flex items-center gap-3">
                <Button
                    nativeButton={false}
                    variant="ghost"
                    size="icon"
                    render={<Link href="/members" />}
                >
                    <ArrowLeft className="h-4 w-4" />
                </Button>

                <h1 className="text-2xl font-bold">
                    Importar membros
                </h1>
            </div>

            <div className="rounded-lg border p-6 space-y-4">
                <h2 className="font-semibold">
                    Arquivo CSV
                </h2>

                <p className="text-sm text-muted-foreground">
                    Faça upload da planilha de membros.
                </p>

                <input
                    id="members-csv"
                    type="file"
                    accept=".csv"
                    onChange={(e) => {
                        const selected = e.target.files?.[0] ?? null;
                        setFile(selected);
                        setError("");
                        setResult(null);
                        setReviewError(null);
                        setReviewNotices([]);
                    }}
                    className="block w-full max-w-lg text-sm file:mr-3 file:rounded-md file:border-0 file:bg-muted file:px-3 file:py-2 file:font-medium hover:file:bg-muted/80"
                />

                {file && (
                    <p className="text-sm text-muted-foreground">
                        Arquivo selecionado: <span className="font-medium text-foreground">{file.name}</span>
                    </p>
                )}

                {error && (
                    <p role="alert" className="text-sm text-destructive">
                        {error}
                    </p>
                )}

                {reviewNotices.map((notice) => (
                    <p
                        key={`${notice.kind}-${notice.row}`}
                        role="status"
                        className={`rounded-md px-3 py-2 text-sm ${
                            notice.kind === "success"
                                ? "bg-green-100 text-green-800"
                                : "bg-muted text-muted-foreground"
                        }`}
                    >
                        {notice.message}
                    </p>
                ))}

                <div className="flex gap-2">
                    <Button
                        variant="outline"
                        onClick={() =>
                            downloadTemplate()
                        }
                    >
                        <Download className="mr-2 h-4 w-4" />
                        Baixar modelo
                    </Button>

                    {result && result.failed > 0 && result.job_id && (
                        <Button
                            variant="outline"
                            onClick={() => downloadErrors(result.job_id!)}
                        >
                            <Download className="mr-2 h-4 w-4" />
                            Baixar planilha de erros
                        </Button>
                    )}

                    <Button
                        onClick={handleImport}
                        disabled={loading}
                    >
                        {loading ? (
                            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        ) : (
                            <Upload className="mr-2 h-4 w-4" />
                        )}

                        Importar
                    </Button>
                </div>

                {result && (
                    <p className="text-sm">
                        Importados: <strong>{result.imported}</strong> · Com erro:{" "}
                        <strong>{result.failed}</strong>
                    </p>
                )}
            </div>
            {
                importErrors.length > 0 && (
                    <div className="rounded-lg border border-amber-200 p-6">
                        <h3 className="font-semibold text-amber-700">
                            Linhas para revisão
                        </h3>

                        <ul className="mt-3 space-y-3">
                            {importErrors.map((err) => (
                                <li
                                    key={err.row}
                                    className="rounded-md border p-4 text-sm"
                                >
                                    <p className="font-medium">
                                        Linha {err.row}: {err.error}
                                    </p>

                                    {err.code === "possible_duplicate" && (
                                        <Button
                                            className="mt-3"
                                            size="sm"
                                            variant="outline"
                                            onClick={() => setReviewError(err)}
                                        >
                                            Comparar membros
                                        </Button>
                                    )}
                                </li>
                            ))}
                        </ul>
                    </div>
                )
            }

            <Dialog
                open={reviewError !== null}
                onOpenChange={(open) => {
                    if (!open && confirmingRow === null && dismissingRow === null) {
                        setReviewError(null);
                    }
                }}
            >
                <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-4xl">
                    <DialogHeader>
                        <DialogTitle className="flex items-center gap-2">
                            <AlertTriangle className="h-5 w-5 text-amber-600" />
                            Possível membro duplicado
                        </DialogTitle>
                        <DialogDescription>
                            Compare a linha {reviewError?.row} da planilha com o cadastro existente.
                        </DialogDescription>
                    </DialogHeader>

                    <div className="space-y-4">
                        {reviewError?.candidates?.map((candidate) => (
                            <div key={candidate.member_id} className="space-y-3 rounded-lg border p-4">
                                <div className="grid gap-4 md:grid-cols-2">
                                    <MemberComparisonCard
                                        title="Membro da planilha"
                                        member={memberFromImportRow(reviewError.data)}
                                    />
                                    <MemberComparisonCard
                                        title="Membro já cadastrado"
                                        member={{
                                            name: candidate.name,
                                            phone: formatBrazilianPhone(candidate.phone),
                                            congregation: candidate.congregation,
                                            address: candidate.address,
                                            addressNumber: candidate.address_number,
                                            neighborhood: candidate.neighborhood,
                                            city: candidate.city,
                                        }}
                                    />
                                </div>
                                <p className="text-xs text-muted-foreground">
                                    Dados correspondentes: {candidate.matched_fields
                                        .map((field) => duplicateFieldLabels[field] || field)
                                        .join(", ")}
                                </p>
                            </div>
                        ))}
                    </div>

                    <DialogFooter>
                        <Button
                            variant="outline"
                            onClick={() => reviewError && dismissDuplicate(reviewError.row)}
                            disabled={confirmingRow !== null || dismissingRow !== null}
                        >
                            {dismissingRow !== null && (
                                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                            )}
                            Cancelar
                        </Button>
                        <Button
                            onClick={() => reviewError && confirmDuplicate(reviewError.row)}
                            disabled={confirmingRow !== null || dismissingRow !== null}
                        >
                            {confirmingRow !== null && (
                                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                            )}
                            Cadastrar mesmo assim
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div >
    );
}

interface ComparisonMember {
    name: string;
    phone: string;
    congregation: string;
    address: string;
    addressNumber: string;
    neighborhood: string;
    city: string;
}

function memberFromImportRow(row?: string[]): ComparisonMember {
    return {
        name: row?.[0] || "",
        phone: formatBrazilianPhone(row?.[1] || ""),
        congregation: row?.[8] || "",
        address: row?.[14] || "",
        addressNumber: row?.[15] || "",
        neighborhood: row?.[17] || "",
        city: row?.[18] || "",
    };
}

function MemberComparisonCard({
    title,
    member,
}: {
    title: string;
    member: ComparisonMember;
}) {
    const fields = [
        ["Nome", member.name],
        ["Telefone", member.phone],
        ["Congregação", member.congregation],
        ["Endereço", member.address],
        ["Número", member.addressNumber],
        ["Bairro", member.neighborhood],
        ["Cidade", member.city],
    ];

    return (
        <div className="rounded-lg bg-muted/50 p-4">
            <h4 className="mb-3 font-semibold">{title}</h4>
            <dl className="space-y-2">
                {fields.map(([label, value]) => (
                    <div key={label} className="grid grid-cols-[7rem_1fr] gap-2 text-sm">
                        <dt className="text-muted-foreground">{label}</dt>
                        <dd className="font-medium">{value || "—"}</dd>
                    </div>
                ))}
            </dl>
        </div>
    );
}
