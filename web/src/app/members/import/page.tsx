"use client";

import { useState } from "react";
import Link from "next/link";
import { ArrowLeft, Upload, Download, Loader2 } from "lucide-react";

import { Button } from "@/components/ui/button";
import { useAuth } from "@/lib/auth";

export default function ImportMembersPage() {
    const { token } = useAuth();

    const [file, setFile] = useState<File | null>(null);
    const [loading, setLoading] = useState(false);
    const [result, setResult] = useState<any>(null);

    async function handleImport() {
        console.log("IMPORT CLICK");
        if (!file) return;

        setLoading(true);

        try {
            const formData = new FormData();
            formData.append("file", file);

            const res = await fetch(
                "http://localhost:8080/api/members/import",
                {
                    method: "POST",
                    headers: {
                        Authorization: `Bearer ${token}`,
                    },
                    body: formData,
                }
            );

            const json = await res.json();

            setResult(json.data);
        } catch (err) {
            console.error(err);
            alert("Erro ao importar arquivo");
        } finally {
            setLoading(false);
        }
    }

    async function downloadErrors(jobId: string) {
        try {
            const res = await fetch(
                `http://localhost:8080/api/members/import/errors/${jobId}`,
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
            "http://localhost:8080/api/members/import/template",
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

    return (
        <div className="space-y-6">
            <div className="flex items-center gap-3">
                <Link href="/members">
                    <Button
                        variant="ghost"
                        size="icon"
                    >
                        <ArrowLeft className="h-4 w-4" />
                    </Button>
                </Link>

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
                    type="file"
                    accept=".csv"
                    onChange={(e) => {

                        const selected = e.target.files?.[0] ?? null;

                        console.log(selected);

                        setFile(selected);
                    }}
                />

                <Button
                    onClick={handleImport}
                    disabled={!file || loading}
                >
                    {loading ? (
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    ) : (
                        <Upload className="mr-2 h-4 w-4" />
                    )}

                    Importar
                </Button>

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

                    {result?.failed > 0 && result?.job_id && (
                        <Button
                            variant="outline"
                            onClick={() => downloadErrors(result.job_id)}
                        >
                            <Download className="mr-2 h-4 w-4" />
                            Baixar planilha de erros
                        </Button>
                    )}

                    <Button
                        onClick={handleImport}
                        disabled={!file || loading}
                    >
                        {loading ? (
                            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        ) : (
                            <Upload className="mr-2 h-4 w-4" />
                        )}

                        Importar
                    </Button>
                </div>
            </div>
            {
                result?.errors?.length > 0 && (
                    <div className="rounded-lg border border-red-200 p-6">
                        <h3 className="font-semibold text-red-600">
                            Linhas com erro
                        </h3>

                        <ul className="mt-3 space-y-2">
                            {result.errors.map((err: any, index: number) => (
                                <li
                                    key={index}
                                    className="text-sm text-red-600"
                                >
                                    Linha {err.row}: {err.error}
                                </li>
                            ))}
                        </ul>
                    </div>
                )
            }
        </div >
    );
}
