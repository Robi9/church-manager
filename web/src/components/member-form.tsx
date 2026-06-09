"use client";

import { useState, type FormEvent } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/lib/auth";
import { api } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
    Card,
    CardContent,
    CardHeader,
    CardTitle,
} from "@/components/ui/card";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
import { ArrowLeft, Loader2, Check } from "lucide-react";
import Link from "next/link";

interface MemberFormProps {
    mode: "create" | "edit";
    initialData?: any;
}

export function MemberForm({
    mode,
    initialData,
}: MemberFormProps) {
    const { token } = useAuth();
    const router = useRouter();
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState("");

    const [form, setForm] = useState({
        name: initialData?.name || "",
        email: initialData?.email || "",
        phone: initialData?.phone || "",
        status: initialData?.status || "active",

        member_since: initialData?.member_since?.slice(0, 10) || "",

        church_role: initialData?.church_role || "",

        marital_status:
            initialData?.marital_status || "single",

        origin_denomination:
            initialData?.origin_denomination || "",

        baptized: initialData?.baptized || false,

        baptism_date:
            initialData?.baptism_date?.slice(0, 10) || "",

        contacted: initialData?.contacted || false,

        contacted_at:
            initialData?.contacted_at?.slice(0, 10) || "",

        membership_course_completed:
            initialData?.membership_course_completed || false,

        membership_course_completed_at:
            initialData?.membership_course_completed_at?.slice(0, 10) || "",
    });

    function update(field: string, value: string | boolean | null) {
        setForm((prev) => ({ ...prev, [field]: value }));
    }

    async function handleSubmit(e: FormEvent) {
        e.preventDefault();
        setError("");
        setLoading(true);

        const body = {
            ...form,
            member_since: form.member_since ? new Date(form.member_since).toISOString() : null,
            baptism_date: form.baptism_date ? new Date(form.baptism_date).toISOString() : null,
            contacted_at: form.contacted_at ? new Date(form.contacted_at).toISOString() : null,
            membership_course_completed_at: form.membership_course_completed_at
                ? new Date(form.membership_course_completed_at).toISOString()
                : null,
        };

        try {
            if (mode === "create") {
                await api("/members", {
                    method: "POST",
                    body,
                    token,
                });
            } else {
                await api(`/members/${initialData.id}`, {
                    method: "PUT",
                    body,
                    token,
                });
            }
            router.push("/members");
        } catch (err) {
            setError(err instanceof Error ? err.message : "Ocorreu um erro");
        } finally {
            setLoading(false);
        }
    }

    return (
        <div className="mx-auto max-w-2xl space-y-6">
            <div className="flex items-center gap-3">
                <Button variant="ghost" size="icon" render={<Link href="/members" />}>
                    <ArrowLeft className="h-4 w-4" />
                </Button>
                <h1 className="text-2xl font-bold">
                    {mode === "create" ? "Novo membro" : "Editar membro"}
                </h1>
            </div>

            <form onSubmit={handleSubmit} className="space-y-6">
                {error && (
                    <div className="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">
                        {error}
                    </div>
                )}

                {/* Dados pessoais */}
                <Card>
                    <CardHeader>
                        <CardTitle className="text-lg">Dados pessoais</CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="space-y-2">
                            <Label htmlFor="name">Nome *</Label>
                            <Input
                                id="name"
                                placeholder="Nome completo"
                                value={form.name}
                                onChange={(e) => update("name", e.target.value)}
                                required
                            />
                        </div>

                        <div className="grid gap-4 sm:grid-cols-2">
                            <div className="space-y-2">
                                <Label htmlFor="email">Email</Label>
                                <Input
                                    id="email"
                                    type="email"
                                    placeholder="email@exemplo.com"
                                    value={form.email}
                                    onChange={(e) => update("email", e.target.value)}
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="phone">Telefone</Label>
                                <Input
                                    id="phone"
                                    placeholder="(00) 00000-0000"
                                    value={form.phone}
                                    onChange={(e) => update("phone", e.target.value)}
                                />
                            </div>
                        </div>

                        <div className="grid gap-4 sm:grid-cols-2">
                            <div className="space-y-2">
                                <Label>Estado civil</Label>
                                <Select
                                    value={form.marital_status}
                                    onValueChange={(v) => update("marital_status", v)}
                                >
                                    <SelectTrigger>
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="single">Solteiro(a)</SelectItem>
                                        <SelectItem value="married">Casado(a)</SelectItem>
                                        <SelectItem value="divorced">Divorciado(a)</SelectItem>
                                        <SelectItem value="widowed">Viúvo(a)</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>
                            <div className="space-y-2">
                                <Label>Status</Label>
                                <Select
                                    value={form.status}
                                    onValueChange={(v) => update("status", v)}
                                >
                                    <SelectTrigger>
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="active">Ativo</SelectItem>
                                        <SelectItem value="inactive">Inativo</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>
                        </div>
                    </CardContent>
                </Card>

                {/* Dados da igreja */}
                <Card>
                    <CardHeader>
                        <CardTitle className="text-lg">Dados da igreja</CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="grid gap-4 sm:grid-cols-2">
                            <div className="space-y-2">
                                <Label htmlFor="church_role">Função na igreja</Label>
                                <Input
                                    id="church_role"
                                    placeholder="Ex: Líder de louvor"
                                    value={form.church_role}
                                    onChange={(e) => update("church_role", e.target.value)}
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="origin_denomination">Denominação de origem</Label>
                                <Input
                                    id="origin_denomination"
                                    placeholder="Ex: Assembleia de Deus"
                                    value={form.origin_denomination}
                                    onChange={(e) => update("origin_denomination", e.target.value)}
                                />
                            </div>
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="member_since">Membro desde</Label>
                            <Input
                                id="member_since"
                                type="date"
                                value={form.member_since}
                                onChange={(e) => update("member_since", e.target.value)}
                            />
                        </div>

                        <Separator />

                        {/* Batismo */}
                        <div className="space-y-3">
                            <CheckboxItem
                                checked={form.baptized}
                                onChange={(v) => update("baptized", v)}
                                label="Batizado(a)"
                            />
                            {form.baptized && (
                                <div className="ml-8 space-y-2">
                                    <Label htmlFor="baptism_date">Data do batismo</Label>
                                    <Input
                                        id="baptism_date"
                                        type="date"
                                        value={form.baptism_date}
                                        onChange={(e) => update("baptism_date", e.target.value)}
                                    />
                                </div>
                            )}
                        </div>

                        <Separator />

                        {/* Contato */}
                        <div className="space-y-3">
                            <CheckboxItem
                                checked={form.contacted}
                                onChange={(v) => update("contacted", v)}
                                label="Contatado(a)"
                            />
                            {form.contacted && (
                                <div className="ml-8 space-y-2">
                                    <Label htmlFor="contacted_at">Data do contato</Label>
                                    <Input
                                        id="contacted_at"
                                        type="date"
                                        value={form.contacted_at}
                                        onChange={(e) => update("contacted_at", e.target.value)}
                                    />
                                </div>
                            )}
                        </div>

                        <Separator />

                        {/* Curso de membresia */}
                        <div className="space-y-3">
                            <CheckboxItem
                                checked={form.membership_course_completed}
                                onChange={(v) => update("membership_course_completed", v)}
                                label="Curso de membresia concluído"
                            />
                            {form.membership_course_completed && (
                                <div className="ml-8 space-y-2">
                                    <Label htmlFor="membership_course_completed_at">
                                        Data de conclusão
                                    </Label>
                                    <Input
                                        id="membership_course_completed_at"
                                        type="date"
                                        value={form.membership_course_completed_at}
                                        onChange={(e) =>
                                            update("membership_course_completed_at", e.target.value)
                                        }
                                    />
                                </div>
                            )}
                        </div>
                    </CardContent>
                </Card>

                <div className="flex justify-end gap-3">
                    <Button variant="outline" type="button" render={<Link href="/members" />}>
                        Cancelar
                    </Button>
                    <Button type="submit" disabled={loading}>
                        {loading ? (
                            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        ) : (
                            <Check className="mr-2 h-4 w-4" />
                        )}
                        {mode === "create"
                            ? "Salvar membro"
                            : "Atualizar membro"}
                    </Button>
                </div>
            </form>
        </div>
    );
}

function CheckboxItem({
    checked,
    onChange,
    label,
}: {
    checked: boolean;
    onChange: (v: boolean) => void;
    label: string;
}) {
    return (
        <label className="flex cursor-pointer items-center gap-3">
            <button
                type="button"
                role="checkbox"
                aria-checked={checked}
                onClick={() => onChange(!checked)}
                className={`flex h-5 w-5 shrink-0 items-center justify-center rounded border transition-colors ${checked
                    ? "border-primary bg-primary text-primary-foreground"
                    : "border-input bg-background"
                    }`}
            >
                {checked && <Check className="h-3 w-3" />}
            </button>
            <span className="text-sm">{label}</span>
        </label>
    );
}
