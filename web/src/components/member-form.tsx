"use client";

import { useState, type FormEvent } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/lib/auth";
import { api, ApiError } from "@/lib/api";
import {
    duplicateFieldLabels,
    duplicateRiskLabels,
    type DuplicateCheckResult,
} from "@/lib/member-duplicates";
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
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { ArrowLeft, Loader2, Check, AlertTriangle, Eye } from "lucide-react";
import Link from "next/link";

interface MemberFormProps {
    mode: "create" | "edit";
    initialData?: MemberFormInitialData;
}

interface MemberFormInitialData {
    id: number;
    name?: string;
    phone?: string;
    status?: string;
    member_since?: string | null;
    church_role?: string;
    marital_status?: string;
    origin_denomination?: string;
    congregation?: string;
    address?: string;
    address_number?: string;
    address_complement?: string;
    neighborhood?: string;
    city?: string;
    state?: string;
    baptized?: boolean;
    baptism_date?: string | null;
    contacted?: boolean;
    contacted_at?: string | null;
    membership_course_completed?: boolean;
    membership_course_completed_at?: string | null;
}

export function MemberForm({
    mode,
    initialData,
}: MemberFormProps) {
    const { token } = useAuth();
    const router = useRouter();
    const memberID = initialData?.id;
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState("");
    const [duplicateResult, setDuplicateResult] = useState<DuplicateCheckResult | null>(null);
    const [pendingBody, setPendingBody] = useState<Record<string, unknown> | null>(null);

    const [form, setForm] = useState({
        name: initialData?.name || "",
        phone: initialData?.phone || "",
        status: initialData?.status || "active",

        member_since: initialData?.member_since?.slice(0, 10) || "",

        church_role: initialData?.church_role || "",

        marital_status:
            initialData?.marital_status || "single",

        origin_denomination:
            initialData?.origin_denomination || "",

        congregation: initialData?.congregation || "Sede",

        address: initialData?.address || "",
        address_number: initialData?.address_number || "",
        address_complement: initialData?.address_complement || "",
        neighborhood: initialData?.neighborhood || "",
        city: initialData?.city || "",
        state: initialData?.state || "",

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
            const check = await api<DuplicateCheckResult>("/members/check-duplicates", {
                method: "POST",
                body: {
                    ...body,
                    exclude_member_id: mode === "edit" ? memberID : undefined,
                },
                token,
            });

            if (check.data.has_possible_duplicates) {
                setPendingBody(body);
                setDuplicateResult(check.data);
                return;
            }

            await persistMember(body, false);
        } catch (err) {
            handleRequestError(err, body);
        } finally {
            setLoading(false);
        }
    }

    async function persistMember(body: Record<string, unknown>, forceCreate: boolean) {
        if (mode === "create") {
            await api("/members", {
                method: "POST",
                body: { ...body, force_create: forceCreate },
                token,
            });
        } else {
            if (!memberID) {
                throw new Error("Membro inválido para edição");
            }
            await api(`/members/${memberID}`, {
                method: "PUT",
                body: { ...body, force_create: forceCreate },
                token,
            });
        }
        router.push("/members");
    }

    function handleRequestError(err: unknown, body: Record<string, unknown>) {
        if (
            err instanceof ApiError &&
            err.status === 409 &&
            isDuplicateResult(err.data)
        ) {
            setPendingBody(body);
            setDuplicateResult(err.data);
            return;
        }
        setError(err instanceof Error ? err.message : "Ocorreu um erro");
    }

    async function confirmDuplicate() {
        if (!pendingBody) {
            return;
        }
        setError("");
        setLoading(true);
        try {
            await persistMember(pendingBody, true);
            setDuplicateResult(null);
        } catch (err) {
            handleRequestError(err, pendingBody);
        } finally {
            setLoading(false);
        }
    }

    function cancelDuplicate() {
        setDuplicateResult(null);
        setPendingBody(null);
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

                        <div className="space-y-2">
                            <Label htmlFor="phone">Telefone</Label>
                            <Input
                                id="phone"
                                placeholder="(00) 00000-0000"
                                value={form.phone}
                                onChange={(e) => update("phone", e.target.value)}
                            />
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

                        <div className="space-y-2">
                            <Label>Congregação</Label>
                            <Select
                                value={form.congregation}
                                onValueChange={(value) => update("congregation", value)}
                            >
                                <SelectTrigger>
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    {[
                                        "Sede",
                                        "Várzea de Cima",
                                        "Caraúno",
                                        "Cohab",
                                        "Quixadá",
                                        "Fortaleza",
                                        "Castelo",
                                        "Conjunto Esperança",
                                    ].map((congregation) => (
                                        <SelectItem key={congregation} value={congregation}>
                                            {congregation}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
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

                <Card>
                    <CardHeader>
                        <CardTitle className="text-lg">Endereço</CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="grid gap-4 sm:grid-cols-[1fr_8rem]">
                            <div className="space-y-2">
                                <Label htmlFor="address">Logradouro</Label>
                                <Input
                                    id="address"
                                    placeholder="Rua, avenida..."
                                    value={form.address}
                                    onChange={(e) => update("address", e.target.value)}
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="address_number">Número</Label>
                                <Input
                                    id="address_number"
                                    value={form.address_number}
                                    onChange={(e) => update("address_number", e.target.value)}
                                />
                            </div>
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="address_complement">Complemento</Label>
                            <Input
                                id="address_complement"
                                value={form.address_complement}
                                onChange={(e) => update("address_complement", e.target.value)}
                            />
                        </div>

                        <div className="grid gap-4 sm:grid-cols-2">
                            <div className="space-y-2">
                                <Label htmlFor="neighborhood">Bairro</Label>
                                <Input
                                    id="neighborhood"
                                    value={form.neighborhood}
                                    onChange={(e) => update("neighborhood", e.target.value)}
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="city">Cidade</Label>
                                <Input
                                    id="city"
                                    value={form.city}
                                    onChange={(e) => update("city", e.target.value)}
                                />
                            </div>
                        </div>

                        <div className="max-w-32 space-y-2">
                            <Label htmlFor="state">Estado</Label>
                            <Input
                                id="state"
                                maxLength={2}
                                placeholder="CE"
                                value={form.state}
                                onChange={(e) => update("state", e.target.value.toUpperCase())}
                            />
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

            <Dialog
                open={duplicateResult !== null}
                onOpenChange={(open) => {
                    if (!open) cancelDuplicate();
                }}
            >
                <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-2xl">
                    <DialogHeader>
                        <DialogTitle className="flex items-center gap-2">
                            <AlertTriangle className="h-5 w-5 text-amber-600" />
                            Possíveis membros duplicados
                        </DialogTitle>
                        <DialogDescription>
                            Encontramos registros semelhantes. Confira antes de continuar; nenhum
                            membro será mesclado automaticamente.
                        </DialogDescription>
                    </DialogHeader>

                    <div className="space-y-3">
                        {duplicateResult?.candidates.map((candidate) => (
                            <div key={candidate.member_id} className="rounded-lg border p-4">
                                <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                                    <div>
                                        <p className="font-semibold">{candidate.name}</p>
                                        <p className="text-sm text-muted-foreground">
                                            {candidate.phone || "Telefone não informado"}
                                            {candidate.congregation
                                                ? ` · ${candidate.congregation}`
                                                : ""}
                                        </p>
                                    </div>
                                    <span
                                        className={`w-fit rounded-full px-2.5 py-1 text-xs font-medium ${
                                            candidate.risk === "high"
                                                ? "bg-red-100 text-red-700"
                                                : "bg-amber-100 text-amber-700"
                                        }`}
                                    >
                                        {duplicateRiskLabels[candidate.risk]}
                                    </span>
                                </div>

                                <p className="mt-3 text-xs font-medium text-muted-foreground">
                                    Dados correspondentes
                                </p>
                                <div className="mt-2 flex flex-wrap gap-2">
                                    {candidate.matched_fields.map((field) => (
                                        <span
                                            key={field}
                                            className="rounded-md bg-muted px-2 py-1 text-xs"
                                        >
                                            {duplicateFieldLabels[field] || field}
                                        </span>
                                    ))}
                                </div>

                                <Button
                                    className="mt-3"
                                    size="sm"
                                    variant="outline"
                                    render={<Link href={`/members/${candidate.member_id}`} />}
                                >
                                    <Eye className="mr-2 h-4 w-4" />
                                    Visualizar membro
                                </Button>
                            </div>
                        ))}
                    </div>

                    <DialogFooter>
                        <Button variant="outline" onClick={cancelDuplicate} disabled={loading}>
                            Cancelar
                        </Button>
                        <Button onClick={confirmDuplicate} disabled={loading}>
                            {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                            {mode === "create" ? "Cadastrar mesmo assim" : "Atualizar mesmo assim"}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
}

function isDuplicateResult(value: unknown): value is DuplicateCheckResult {
    if (!value || typeof value !== "object") {
        return false;
    }
    const result = value as Partial<DuplicateCheckResult>;
    return Array.isArray(result.candidates) && typeof result.has_possible_duplicates === "boolean";
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
