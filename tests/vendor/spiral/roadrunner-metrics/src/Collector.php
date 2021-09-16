<?php

/**
 * This file is part of RoadRunner package.
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

namespace Spiral\RoadRunner\Metrics;

use JetBrains\PhpStorm\ExpectedValues;
use JetBrains\PhpStorm\Pure;
use JsonSerializable;

/**
 * @psalm-import-type CollectorType from CollectorInterface
 * @psalm-import-type ArrayFormatType from CollectorInterface
 */
final class Collector implements CollectorInterface, JsonSerializable
{
    /**
     * @var CollectorType
     */
    #[ExpectedValues(valuesFromClass: self::class)]
    private string $type;

    /**
     * @var string
     */
    private string $namespace = '';

    /**
     * @var string
     */
    private string $subsystem = '';

    /**
     * @var string
     */
    private string $help = '';

    /**
     * @var array<array-key, string>
     */
    private array $labels = [];

    /**
     * @var array<array-key, float>
     */
    private array $buckets = [];

    /**
     * @param CollectorType $type
     */
    private function __construct(
        #[ExpectedValues(valuesFromClass: self::class)]
        string $type
    ) {
        $this->type = $type;
    }

    /**
     * {@inheritDoc}
     */
    #[Pure]
    public function withNamespace(
        string $namespace
    ): self {
        $self = clone $this;
        $self->namespace = $namespace;

        return $self;
    }

    /**
     * {@inheritDoc}
     */
    #[Pure]
    public function withSubsystem(
        string $subsystem
    ): self {
        $self = clone $this;
        $self->subsystem = $subsystem;

        return $self;
    }

    /**
     * {@inheritDoc}
     */
    #[Pure]
    public function withHelp(
        string $help
    ): self {
        $self = clone $this;
        $self->help = $help;

        return $self;
    }

    /**
     * {@inheritDoc}
     */
    #[Pure]
    public function withLabels(
        string ...$label
    ): self {
        $self = clone $this;
        $self->labels = $label;

        return $self;
    }

    /**
     * {@inheritDoc}
     */
    #[Pure]
    public function toArray(): array
    {
        return [
            'namespace' => $this->namespace,
            'subsystem' => $this->subsystem,
            'type'      => $this->type,
            'help'      => $this->help,
            'labels'    => $this->labels,
            'buckets'   => $this->buckets,
        ];
    }

    /**
     * @return ArrayFormatType
     */
    #[Pure]
    public function jsonSerialize(): array
    {
        return $this->toArray();
    }

    /**
     * New histogram metric.
     *
     * @param float ...$bucket
     * @return static
     */
    #[Pure]
    public static function histogram(float ...$bucket): self
    {
        $self = new self(self::TYPE_HISTOGRAM);
        /** @psalm-suppress ImpurePropertyAssignment */
        $self->buckets = $bucket;

        return $self;
    }

    /**
     * New gauge metric.
     *
     * @return static
     */
    #[Pure]
    public static function gauge(): self
    {
        return new self(self::TYPE_GAUGE);
    }

    /**
     * New counter metric.
     *
     * @return static
     */
    #[Pure]
    public static function counter(): self
    {
        return new self(self::TYPE_COUNTER);
    }
}
